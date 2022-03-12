package server

import (
	"compress/gzip"
	"github.com/fd239/go_url_shortener/internal/app/handlers"
	"github.com/fd239/go_url_shortener/internal/app/middleware"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
)

type Server interface {
	Start() error
}

type server struct {
	address string
	baseURL string
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// проверяем, что клиент поддерживает gzip-сжатие
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func CreateRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware)
	//r.Use(middleware.DecompressMiddleware)
	r.Get("/api/user/urls", handlers.GetUserURLs)
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
	return http.ListenAndServe(s.address, gzipHandle(r))
}
