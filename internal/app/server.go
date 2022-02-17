package app

import (
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/_const"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"log"
	"net/http"
)

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

	url := SaveShortRoute(shorten.URL)

	response := ShortenResponse{Result: url}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)

}

func getUrl(w http.ResponseWriter, r *http.Request) {
	urlId := chi.URLParam(r, "id")

	url, err := GetShortRoute(urlId)

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
		log.Println(_const.ErrMsg_BodyReadError)
		http.Error(w, _const.ErrMsg_BodyReadError, http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		log.Println(_const.ErrMsg_EmptyBody)
		http.Error(w, _const.ErrMsg_EmptyBody, http.StatusBadRequest)
		return
	}

	shortUrl := SaveShortRoute(string(body))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", _const.Hostname, shortUrl)))
}

func ServerStart() {
	router := CreateRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
