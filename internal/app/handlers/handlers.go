package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/context"
	"golang.org/x/sync/errgroup"
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

func BatchURLs(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

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
		log.Printf("json.Encode: %v\n", err)
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	userID := context.Get(r, "userID")
	batchItemsResponse, batchErr := Store.CreateItems(batchItems, fmt.Sprintf("%v", userID))

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

func DeleteURLs(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("delete urls body read error: %v\n", err)
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		log.Println(common.ErrEmptyBody)
		http.Error(w, common.ErrEmptyBody.Error(), http.StatusBadRequest)
		return
	}

	var deleteIDs []string
	err = json.Unmarshal(body, &deleteIDs)

	if err != nil {
		log.Printf("json.Encode: %v\n", err)
		http.Error(w, common.ErrBodyReadError.Error(), http.StatusBadRequest)
		return
	}

	g, _ := errgroup.WithContext(r.Context())

	g.Go(func() error {
		return Store.UpdateItems(deleteIDs)
	})

	if err = g.Wait(); err != nil {
		http.Error(w, common.ErrResponseEncode.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
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
	body, err := ioutil.ReadAll(r.Body)
	status := 0

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
			status = http.StatusConflict
		} else {
			errString := fmt.Sprintf("Save short route error: %v\n", err)
			log.Println(errString)
			http.Error(w, errString, http.StatusBadRequest)
			return
		}
	} else {
		status = http.StatusCreated
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)

	w.Write([]byte(fmt.Sprintf("%s/%s", common.Cfg.BaseURL, shortURL)))
}

func HandleURL(w http.ResponseWriter, r *http.Request) {
	shorten := ShortenRequest{}
	status := 0

	if err := json.NewDecoder(r.Body).Decode(&shorten); err != nil {
		http.Error(w, common.ErrUnableToFindURL.Error(), http.StatusBadRequest)
		return
	}

	userID := context.Get(r, "userID")
	shortURL, err := Store.Insert(shorten.URL, fmt.Sprintf("%v", userID))

	if err != nil {
		if errors.Is(err, common.ErrOriginalURLConflict) {
			status = http.StatusConflict
		} else {
			errString := fmt.Sprintf("Save short route error: %s", err.Error())
			log.Printf("json.Decode: %v\n", err)
			http.Error(w, errString, http.StatusBadRequest)
			return
		}

	} else {
		status = http.StatusCreated
	}

	response := ShortenResponse{Result: fmt.Sprintf("%s/%s", common.Cfg.BaseURL, shortURL)}

	jsonResponse, jsonErr := json.Marshal(response)

	if jsonErr != nil {
		log.Printf("json.Marshall: %v\n", jsonErr)
		http.Error(w, common.ErrResponseEncode.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Add("Accept", "application/json")
	w.WriteHeader(status)

	w.Write(jsonResponse)
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
