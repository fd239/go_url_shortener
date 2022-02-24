package app

import (
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"log"
	"net/http"
)

var config *common.Config

func CreateRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/shorten", handleUrl)
	r.Get("/{id}", getUrl)
	r.Post("/", saveShortUrl)

	return r
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func handleUrl(w http.ResponseWriter, r *http.Request) {
	shorten := ShortenRequest{}

	if err := json.NewDecoder(r.Body).Decode(&shorten); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, _ := DB.SaveShortRoute(shorten.URL)

	response := ShortenResponse{Result: fmt.Sprintf("%s/%s", common.Cfg.BaseURL, url)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)

}

func getUrl(w http.ResponseWriter, r *http.Request) {
	urlId := chi.URLParam(r, "id")

	url, err := DB.GetShortRoute(urlId)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func saveShortUrl(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println(common.ErrBodyReadError)
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		log.Println(common.ErrEmptyBody)
		http.Error(w, common.ErrEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	shortUrl, _ := DB.SaveShortRoute(string(body))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", common.Cfg.BaseURL, shortUrl)))
}

func ServerStart() {
	router := CreateRouter()
	log.Fatal(http.ListenAndServe(common.Cfg.ServerAddress, router))
}
