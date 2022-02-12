package app

import (
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/_const"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"log"
	"net/http"
)

func CreateRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/{id}", getUrl)
	r.Post("/", saveShortUrl)

	return r
}

func getUrl(w http.ResponseWriter, r *http.Request) {
	urlId := chi.URLParam(r, "id")

	url, err := GetShortRoute(urlId)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)

	} else {
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func saveShortUrl(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	if len(body) == 0 {
		log.Println(_const.ErrMsg_EmptyBody)
		http.Error(w, _const.ErrMsg_EmptyBody, http.StatusBadRequest)
		return
	}

	shortUrl := SaveShortRoute(string(body))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(_const.Hostname + shortUrl)))
}

func ServerStart() {
	router := CreateRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
