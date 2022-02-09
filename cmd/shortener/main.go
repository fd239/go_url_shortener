package main

import (
	"log"
	"net/http"
)

func main() {
	urlMap := make(map[string]string)
	http.HandleFunc("/", ShortenerHandler(urlMap))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
