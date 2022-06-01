package handlers

import (
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/server"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExampleSaveShortURL() {
	w := httptest.NewRecorder()
	router := server.CreateRouter()

	router.HandleFunc("/", SaveShortURL)
	Store, _ = storage.InitDB()

	r, _ := http.NewRequest("POST", "/", strings.NewReader(common.TestURL))
	router.ServeHTTP(w, r)
	result := w.Result()
	result.Body.Close()
}
