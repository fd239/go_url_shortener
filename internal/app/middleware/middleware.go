package middleware

import (
	"compress/gzip"
	"expvar"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/crypt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// AuthMiddleware auth to service by token in cookie
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decryptedUserID := ""
		if tokenCookie, err := r.Cookie("token"); err == nil {
			decryptedUserID, err = crypt.Decrypt(tokenCookie.Value)
			if err != nil {
				log.Printf("Decrypt error: %v", err)
				http.Error(w, common.ErrUserCookie.Error(), http.StatusInternalServerError)
				return
			}
		}
		if len(decryptedUserID) == 0 {
			decryptedUserID = uuid.NewString()
			encryptedUserID, err := crypt.Encrypt(decryptedUserID)
			if err != nil {
				log.Printf("Crypt new user encrypt error: %v", err)
				http.Error(w, common.ErrUserCookie.Error(), http.StatusInternalServerError)
				return
			}

			cookie := &http.Cookie{
				Name:   "token",
				Value:  encryptedUserID,
				MaxAge: 300,
			}

			http.SetCookie(w, cookie)

		}
		context.Set(r, "userID", decryptedUserID)
		next.ServeHTTP(w, r)
	})
}

// DecompressMiddleware compressing and decompressing requests and responses
func DecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Printf("gzip body decode error: %v\n", err)
				http.Error(w, common.ErrGzipRead.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gz
			defer gz.Close()

		}

		if r.Header.Get(`Accept-Encoding`) != `gzip` {
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

// Profiler chi default profiler handler
func Profiler() http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.HandleFunc("/vars", expVars)

	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/pprof/heap", pprof.Handler("heap"))
	r.Handle("/pprof/block", pprof.Handler("block"))
	r.Handle("/pprof/allocs", pprof.Handler("allocs"))

	return r
}

// Replicated from expvar.go as not public.
func expVars(w http.ResponseWriter, r *http.Request) {
	first := true
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}
