package server

import (
	"context"
	"github.com/fd239/go_url_shortener/internal/app/handlers"
	"github.com/fd239/go_url_shortener/internal/app/middleware"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

const (
	certFile = "ssl/server.crt"
	keyFile  = "ssl/ca.key"
)

type Server interface {
	Start() error
}

type server struct {
	address string
	baseURL string
	useTls  bool
}

func CreateRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.AuthMiddleware)
	r.Use(middleware.DecompressMiddleware)
	r.Mount("/debug", middleware.Profiler())
	r.Get("/ping", handlers.Ping)
	r.Get("/api/user/urls", handlers.GetUserURLs)
	r.Delete("/api/user/urls", handlers.DeleteURLs)
	r.Post("/api/shorten/batch", handlers.BatchURLs)
	r.Post("/api/shorten", handlers.HandleURL)
	r.Get("/{id}", handlers.GetURL)
	r.Post("/", handlers.SaveShortURL)

	return r
}

// NewServer creating server instance and initialize store
func NewServer(address string, baseURL string, useTls bool) (*server, error) {
	var err error
	handlers.Store, err = storage.InitDB()
	if err != nil {
		return nil, err
	}
	return &server{
		address: address,
		baseURL: baseURL,
		useTls:  useTls,
	}, nil
}

// Start router create and server start
func (s *server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := CreateRouter()

	srv := &http.Server{
		Addr:    ":443",
		Handler: r,
	}

	if s.useTls {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(s.address),
		}

		srv.TLSConfig = manager.TLSConfig()
	}

	if s.useTls {
		go log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
	} else {
		go log.Fatal(srv.ListenAndServe())
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case v := <-quit:
		log.Printf("signal.Notify: %v", v)
	}

	if handlers.Store.PGConn != nil {
		if err := handlers.Store.PGConn.Close(); err != nil {
			log.Printf("Postgres close error: %v", err)
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	return nil
}
