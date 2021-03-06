package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fd239/go_url_shortener/config"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/context"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

var Store *storage.Database

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type trustedSubnetResponse struct {
	Users int `json:"users"`
	Urls  int `json:"urls"`
}

// BatchURLs save multiple urls to storage
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

// DeleteURLs mark url as deleted in Postgres
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

// GetURL GET method for receive url by short id
func GetURL(w http.ResponseWriter, r *http.Request) {
	status := 0

	urlID := chi.URLParam(r, "id")
	url, err := Store.Get(urlID)

	if err != nil {
		if errors.Is(err, common.ErrURLDeleted) {
			status = http.StatusGone
		} else {
			log.Printf("Store GET error: %v\n", err)
			http.Error(w, common.ErrUnableToFindURL.Error(), http.StatusBadRequest)
			return
		}
	} else {
		status = http.StatusTemporaryRedirect
		w.Header().Set("Location", url)
	}

	w.WriteHeader(status)
}

// GetUserURLs Get all user saved urls by user ID
func GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := context.Get(r, "userID")
	userURLs, err := Store.GetUserURL(fmt.Sprintf("%v", userID))

	if err != nil {
		http.Error(w, "store error", http.StatusNoContent)
		return
	}

	if len(userURLs) == 0 {
		http.Error(w, common.ErrNoUserURLs.Error(), http.StatusNoContent)
		return
	}

	var baseURLItems []*storage.UserItem
	for _, v := range userURLs {
		baseURLItems = append(baseURLItems, &storage.UserItem{OriginalURL: v.OriginalURL, ShortURL: fmt.Sprintf("%s/%s", config.Cfg.BaseURL, v.ShortURL)})
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

// SaveShortURL receive short URL in POST method and save it to storage
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

	w.Write([]byte(fmt.Sprintf("%s/%s", config.Cfg.BaseURL, shortURL)))
}

// HandleURL save short URL received from POST request
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

	response := ShortenResponse{Result: fmt.Sprintf("%s/%s", config.Cfg.BaseURL, shortURL)}

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

// Ping short url microservice health check
func Ping(w http.ResponseWriter, _ *http.Request) {
	err := Store.Ping()

	if err != nil {
		log.Printf("DB ping error: %v\n", err)
		http.Error(w, common.ErrPing.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetStats(w http.ResponseWriter, r *http.Request) {
	trustedSubnet := config.Cfg.TrustedSubnet

	if trustedSubnet == "" {
		log.Println("Trusted Subnet not specified")
		http.Error(w, "Trusted Subnet not specified", http.StatusForbidden)
		return
	}
	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		log.Println("Can't parse CIDR")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	ip, err := getRequestIp(r)
	if err != nil {
		log.Println("Can't parse IP")
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if !ipNet.Contains(ip) {
		log.Println("Can't contain IP" + ip.String())
		http.Error(w, "Can't contain IP"+ip.String(), http.StatusForbidden)
		return
	}

	urls := Store.URLCount()
	users := Store.UserCount()

	response := trustedSubnetResponse{
		Users: users,
		Urls:  urls,
	}

	b, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Add("Accept", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write(b)
}

func getRequestIp(r *http.Request) (net.IP, error) {
	remoteAddr := r.RemoteAddr

	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return nil, err
	}

	remoteIp := net.ParseIP(ip)

	realIp := r.Header.Get("X-Real-IP")
	parseIp := net.ParseIP(realIp)

	if parseIp == nil {
		frwIps := r.Header.Get("X-Forwarded-For")
		splitIps := strings.Split(frwIps, ",")
		ip = splitIps[0]
		parseIp = net.ParseIP(ip)
	}

	if remoteIp.Equal(parseIp) {
		return remoteIp, nil
	}

	return nil, errors.New("no ip")
}
