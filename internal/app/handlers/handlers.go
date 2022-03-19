package handlers

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var Store *storage.Database

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func DecompressMiddleware(r *http.Request) io.Reader {
	var reader io.Reader

	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}
	return reader
}

func BatchURLs(w http.ResponseWriter, r *http.Request) {
	reader := DecompressMiddleware(r)
	body, err := ioutil.ReadAll(reader)

	if err != nil {
		log.Printf("batch urls body read error: %v\n", err)
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		log.Println(common.ErrEmptyBody)
		http.Error(w, common.ErrEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	var batchItems []storage.BatchItemRequest
	err = json.Unmarshal(body, &batchItems)

	if err != nil {
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	userID := context.Get(r, "userID")
	batchItemsResponse, batchErr := Store.BatchItems(batchItems, fmt.Sprintf("%v", userID))

	if batchErr != nil {
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(batchItemsResponse); err != nil {
		log.Printf("json.Encode: %v\n", err)
		http.Error(w, common.ErrResponseEncode.Error(), http.StatusBadRequest)
	}

}

func HandleURL(w http.ResponseWriter, r *http.Request) {
	shorten := ShortenRequest{}
	reader := DecompressMiddleware(r)

	if err := json.NewDecoder(reader).Decode(&shorten); err != nil {
		http.Error(w, common.ErrUnableToFindURL.Error(), http.StatusBadRequest)
		return
	}

	userID := context.Get(r, "userID")
	url, err := Store.Insert(shorten.URL, fmt.Sprintf("%v", userID))

	if err != nil {
		errString := fmt.Sprintf("Save short route error: %s", err.Error())
		log.Printf("json.Decode: %v\n", err)
		http.Error(w, errString, http.StatusBadRequest)
		return
	}

	response := ShortenResponse{Result: fmt.Sprintf("%s/%s", common.Cfg.BaseURL, url)}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("json.Encode: %v\n", err)
		http.Error(w, common.ErrResponseEncode.Error(), http.StatusBadRequest)
	}

}

func GetURL(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")

	url, err := Store.Get(urlID)

	if err != nil {
		log.Printf("Store GET error: %v\n", err)
		http.Error(w, common.ErrUnableToFindURL.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := context.Get(r, "userID")
	userURLs := Store.GetUserURL(fmt.Sprintf("%v", userID))

	if len(userURLs) == 0 {
		http.Error(w, common.ErrNoUserURLs.Error(), http.StatusNoContent)
		return
	}

	var baseURLItems []*storage.UserItem
	for _, v := range userURLs {
		baseURLItems = append(baseURLItems, &storage.UserItem{OriginalURL: v.OriginalURL, ShortURL: fmt.Sprintf("%s/%s", common.Cfg.BaseURL, v.ShortURL)})
	}

	userURLsJSON, err := json.Marshal(baseURLItems)

	if err != nil {
		log.Printf("user URLs marshall error: %v\n", err)
		http.Error(w, common.ErrNoUserURLs.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Add("Accept", "application/json")
	w.Write(userURLsJSON)
}

func SaveShortURL(w http.ResponseWriter, r *http.Request) {
	reader := DecompressMiddleware(r)
	body, err := ioutil.ReadAll(reader)

	if err != nil {
		log.Println(common.ErrBodyReadError)
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		log.Println(common.ErrEmptyBody)
		http.Error(w, common.ErrEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	userID := context.Get(r, "userID")
	shortURL, err := Store.Insert(string(body), fmt.Sprintf("%v", userID))

	if err != nil {
		if errors.Is(err, common.ErrOriginalURLConflict) {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		errString := fmt.Sprintf("Save short route error: %v\n", err)
		log.Println(errString)
		http.Error(w, errString, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", common.Cfg.BaseURL, shortURL)))
}

func Ping(w http.ResponseWriter, r *http.Request) {
	err := Store.Ping()

	if err != nil {
		log.Printf("DB ping error: %v\n", err)
		http.Error(w, common.ErrPing.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
