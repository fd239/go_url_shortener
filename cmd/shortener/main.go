package main

import (
	"crypto/md5"
	"encoding/json"
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

type PostBody struct {
	Url_string string `json:"url_string"`
}

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
			w.WriteHeader(http.StatusMovedPermanently)
			return
		} else {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

	} else {
		var pBody = PostBody{}
		body, _ := ioutil.ReadAll(r.Body)

		if err := json.Unmarshal(body, &pBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(pBody.Url_string) == 0 {
			http.Error(w, "No url_string in body", http.StatusBadRequest)
			return
		}

		hashString := hash(pBody.Url_string)
		urlMap[hashString] = pBody.Url_string

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(hashString))

	}
}

func main() {
	http.HandleFunc("/", Handler)
	logger.Fatal(http.ListenAndServe(":8079", nil))
}
