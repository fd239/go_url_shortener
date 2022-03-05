package server

import (
	"github.com/fd239/go_url_shortener/internal/app/handlers"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server interface {
	Start() error
}

type server struct {
	address string
	baseURL string
}

func CreateRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/shorten", handlers.HandleURL)
	r.Get("/{id}", handlers.GetURL)
	r.Post("/", handlers.SaveShortURL)

	return r
}

func NewServer(address string, baseURL string) (*server, error) {
	var err error
	handlers.Store, err = storage.InitDB()
	if err != nil {
		return nil, err
	}
	return &server{
		address: address,
		baseURL: baseURL,
	}, nil
}

func (s *server) Start() error {
	r := CreateRouter()
	return http.ListenAndServe(s.address, r)
}
