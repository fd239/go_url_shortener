package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"log"
	"net/http"
)

var Store *storage.Database

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func HandleURL(w http.ResponseWriter, r *http.Request) {
	shorten := ShortenRequest{}

	if err := json.NewDecoder(r.Body).Decode(&shorten); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, err := Store.Insert(shorten.URL)

	if err != nil {
		errString := fmt.Sprintf("Save short route error: %s", err.Error())
		log.Println(errString)
		http.Error(w, errString, http.StatusBadRequest)
	}

	response := ShortenResponse{Result: fmt.Sprintf("%s/%s", common.Cfg.BaseURL, url)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)

}

func GetURL(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")

	url, err := Store.Get(urlID)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func SaveShortURL(w http.ResponseWriter, r *http.Request) {
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

	shortURL, err := Store.Insert(string(body))

	if err != nil {
		errString := fmt.Sprintf("Save short route error: %s", err.Error())
		log.Println(errString)
		http.Error(w, errString, http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", common.Cfg.BaseURL, shortURL)))
}
