package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/api"
	"github.com/fd239/go_url_shortener/config"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"golang.org/x/sync/errgroup"
	"log"
)

type consumer struct {
	api.UnimplementedShortenerServer
}

func NewConsumer() *consumer {
	return &consumer{}
}

// Ping short url microservice health check
func (c *consumer) Ping(_ context.Context, _ *api.PingRequest) (resp *api.PingResponse, err error) {
	err = Store.Ping()

	if err != nil {
		log.Printf("DB ping error: %v\n", err)
		resp.Error = err.Error()
		return
	}

	return
}

func (c *consumer) GetUserUrls(_ context.Context, req *api.GetUserUrlRequest) (resp *api.GetUserUrlResponse, err error) {
	userID := req.UserId
	userURLs, err := Store.GetUserURL(fmt.Sprintf("%v", userID))

	if err != nil {
		resp.Error = err.Error()
		return
	}

	if len(userURLs) == 0 {
		resp.Error = err.Error()
		return
	}

	var baseURLItems []*storage.UserItem
	for _, v := range userURLs {
		baseURLItems = append(baseURLItems, &storage.UserItem{OriginalURL: v.OriginalURL, ShortURL: fmt.Sprintf("%s/%s", config.Cfg.BaseURL, v.ShortURL)})
	}

	userURLsJSON, err := json.Marshal(baseURLItems)

	if err != nil {
		log.Printf("user URLs marshall error: %v\n", err)
		resp.Error = err.Error()
	}

	resp.UserUrls = string(userURLsJSON)
	return
}

func (c *consumer) DeleteUrls(ctx context.Context, req *api.DeleteUrlsRequest) (resp *api.DeleteUrlsResponse, err error) {
	var deleteIDs []string
	err = json.Unmarshal([]byte(req.UrlsDelete), &deleteIDs)

	if err != nil {
		log.Printf("json.Encode: %v\n", err)
		resp.Error = err.Error()
		return
	}

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		return Store.UpdateItems(deleteIDs)
	})

	if err = g.Wait(); err != nil {
		resp.Error = err.Error()
		return
	}

	return
}

func (c *consumer) BatchUrls(_ context.Context, req *api.BatchUrlsRequest) (resp *api.BatchUrlsResponse, err error) {
	var batchItems []storage.BatchItemRequest
	err = json.Unmarshal([]byte(req.BatchItems), &batchItems)

	if err != nil {
		log.Printf("json.Encode: %v\n", err)
		resp.Error = err.Error()
		return
	}

	batchItemsResponse, batchErr := Store.CreateItems(batchItems, fmt.Sprintf("%v", req.UserId))

	if batchErr != nil {
		resp.Error = batchErr.Error()
		return
	}

	b, err := json.Marshal(batchItemsResponse)

	if err != nil {
		resp.Error = err.Error()
		return
	}

	resp.BatchUrls = string(b)

	return
}

func (c *consumer) HandleUrl(_ context.Context, req *api.HandleUrlRequest) (resp *api.HandleUrlResponse, err error) {
	var shorten ShortenRequest
	err = json.Unmarshal([]byte(req.Url), &shorten)

	if err != nil {
		log.Printf("url unmarshal error: %v\n", err)
		resp.Error = err.Error()
		return
	}

	shortURL, err := Store.Insert(shorten.URL, fmt.Sprintf("%v", req.UserId))

	if err != nil {
		errString := fmt.Sprintf("Save short route error: %s", err.Error())
		log.Printf("json.Decode: %v\n", err)
		resp.Error = errString
		return
	}

	response := ShortenResponse{Result: fmt.Sprintf("%s/%s", config.Cfg.BaseURL, shortURL)}

	jsonResponse, err := json.Marshal(response)

	if err != nil {
		log.Printf("json.Marshall: %v\n", err)
		resp.Error = err.Error()
		return
	}

	resp.ShortUrls = string(jsonResponse)
	return
}

func (c *consumer) GetUrl(_ context.Context, req *api.GetUrlRequest) (resp *api.GetUrlResponse, err error) {
	url, err := Store.Get(req.Id)

	if err != nil {
		log.Printf("Store GET error: %v\n", err)
		resp.Error = common.ErrUnableToFindURL.Error()
		return
	}
	resp.ShortUrl = url
	return
}

func (c *consumer) SaveShortUrl(_ context.Context, req *api.SaveShortUrlRequest) (resp *api.SaveShortUrlResponse, err error) {
	shortURL, err := Store.Insert(req.Url, fmt.Sprintf("%v", req.UserID))

	if err != nil {
		errString := fmt.Sprintf("Save short route error: %v\n", err)
		log.Println(errString)
		resp.Error = errString
		return
	}

	resp.ShortUrl = fmt.Sprintf("%s/%s", config.Cfg.BaseURL, shortURL)
	return
}
