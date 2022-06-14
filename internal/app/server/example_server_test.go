package server

import (
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/handlers"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExampleSaveShortURL() {
	w := httptest.NewRecorder()
	router := CreateRouter()
	r, _ := http.NewRequest("POST", "/", strings.NewReader(common.TestURL))

	handlers.Store, _ = storage.InitDB()

	router.ServeHTTP(w, r)
	router.HandleFunc("/", handlers.SaveShortURL)

	result := w.Result()
	defer result.Body.Close()

	b, err := io.ReadAll(result.Body)
	if err != nil {
		log.Fatalf("Example body read error")
	}
	fmt.Printf("Code: %v\n", result.StatusCode)
	fmt.Printf("Short URL: %v\n", fmt.Sprintf("%s/%s", common.Cfg.BaseURL, string(b)))
	// Output:
	// Code: 201
	// Short URL: http://localhost:8080/88d2d0f8fe07c98da23165c7a8a7acae
}

func ExampleGetShortURL() {
	w := httptest.NewRecorder()
	router := CreateRouter()
	r, _ := http.NewRequest("GET", "/"+common.TestShortID, nil)

	handlers.Store, _ = storage.InitDB()
	handlers.Store.Items[common.TestShortID] = common.TestURL

	router.ServeHTTP(w, r)
	router.HandleFunc("/", handlers.GetURL)

	result := w.Result()
	defer result.Body.Close()

	fmt.Printf("Code: %v\n", result.StatusCode)
	fmt.Printf("Short URL: %v\n", result.Header.Get("Location"))
	// Output:
	// Code: 307
	// Short URL: http://cjdr17afeihmk.biz/kdni9/z9womotrbk
}
