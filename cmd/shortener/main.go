package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	logger = log.Default()
	urlMap = make(map[string]string)
)

func hash(s string) string {
	data := []byte(s)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		url := r.URL.String()
		splittedUrl := strings.Split(url, "/")
		res := splittedUrl[len(splittedUrl)-1]
		if len(res) == 0 {
			http.Error(w, "No ID in request", http.StatusBadRequest)
			return
		}

		if val, ok := urlMap[res]; ok {
			w.Header().Set("Location", val)
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			http.Error(w, "", http.StatusBadRequest)
		}

	} else {
		body, _ := ioutil.ReadAll(r.Body)
		hashString := hash(string(body))
		urlMap[hashString] = string(body)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("http://localhost:8080/" + hashString))
	}
}

func main() {
	http.HandleFunc("/", Handler)
	http.HandleFunc("//", Handler)
	logger.Fatal(http.ListenAndServe(":8080", nil))
}
