package main

import (
	"github.com/fd239/go_url_shortener/internal/app"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func ShortenerHandler(urlMap map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			url := r.URL.String()
			splittedUrl := strings.Split(url, "/")
			res := splittedUrl[len(splittedUrl)-1]
			if len(res) == 0 {
				log.Println("No ID in request")
				http.Error(w, "No ID in request", http.StatusBadRequest)
				return
			}

			if val, ok := urlMap[res]; ok {
				w.Header().Set("Location", val)
				w.WriteHeader(http.StatusTemporaryRedirect)
			} else {
				log.Println("No URL in map")
				http.Error(w, "No URL in map", http.StatusBadRequest)
			}

		} else {
			body, _ := ioutil.ReadAll(r.Body)
			if len(body) == 0 {
				log.Println("Empty body")
				http.Error(w, "Empty body", http.StatusBadRequest)
				return
			}
			hashString := app.GetShortLink(string(body))
			urlMap[hashString] = string(body)

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("http://localhost:8080/" + hashString))
		}
	}
}
