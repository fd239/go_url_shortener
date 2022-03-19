package middleware

import (
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/crypt"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	"log"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decryptedUserID := ""
		if tokenCookie, err := r.Cookie("token"); err == nil {
			decryptedUserID, err = crypt.Decrypt(tokenCookie.Value)
			if err != nil {
				log.Println(fmt.Sprintf("Decrypt error: %v", err))
			}
		}
		if len(decryptedUserID) == 0 {
			decryptedUserID = uuid.NewString()
			encryptedUserID, err := crypt.Encrypt(decryptedUserID)
			if err != nil {
				log.Println(fmt.Sprintf("Crypt new user encrypt error: %v", err))
				http.Error(w, common.ErrUserCookie.Error(), http.StatusBadRequest)
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

//func DecompressMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		if r.Header.Get(`Content-Encoding`) == `gzip` {
//			gz, err := gzip.NewReader(r.Body)
//			if err != nil {
//				log.Println(fmt.Sprintf("gzip body decode error: %v", err))
//			}
//			r.Body = gz
//			defer gz.Close()
//
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}
