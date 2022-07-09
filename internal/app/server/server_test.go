package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fd239/go_url_shortener/config"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/handlers"
	"github.com/fd239/go_url_shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func getJSONRequest() *bytes.Buffer {
	var buf bytes.Buffer
	req := handlers.ShortenRequest{URL: common.TestURL}
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		log.Println("JSON encode error")
	}

	return &buf
}

func getUserUrlsRequest() *bytes.Buffer {
	var buf bytes.Buffer
	req := handlers.ShortenRequest{URL: common.TestURL}
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		log.Println("JSON encode error")
	}

	return &buf
}

func getJSONResponse() string {
	res := handlers.ShortenResponse{Result: fmt.Sprintf("%s/%s", config.Cfg.BaseURL, common.TestShortID)}
	b, err := json.Marshal(res)

	if err != nil {
		log.Println("JSON Marshall error: ", err.Error())
		return ""
	}

	return string(b)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string, string, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	location := resp.Header.Get("location")
	contentType := resp.Header.Get("Content-Type")

	stringBody := strings.TrimSuffix(string(respBody), "\n")

	return resp, stringBody, location, contentType
}

func TestRouter(t *testing.T) {
	type want struct {
		code        int
		response    string
		location    string
		contentType string
	}
	type args struct {
		method string
		target string
		body   io.Reader
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "POST 200",
			args: args{http.MethodPost, "/", strings.NewReader(common.TestURL)},
			want: want{http.StatusCreated, fmt.Sprintf("%s/%s", config.Cfg.BaseURL, common.TestShortID), "", "text/plain; charset=utf-8"},
		},
		{
			name: "POST 400 Empty body",
			args: args{http.MethodPost, "/", nil},
			want: want{http.StatusBadRequest, common.ErrEmptyBody.Error(), "", "text/plain; charset=utf-8"},
		},
		{
			name: "GET 307",
			args: args{http.MethodGet, "/" + common.TestShortID, nil},
			want: want{http.StatusTemporaryRedirect, "", common.TestURL, ""},
		},
		{
			name: "GET 405 No ID in request",
			args: args{http.MethodGet, "/", nil},
			want: want{http.StatusMethodNotAllowed, "", "", ""},
		},
		{
			name: "GET 400 No URL in map",
			args: args{http.MethodGet, "/123", nil},
			want: want{http.StatusBadRequest, common.ErrUnableToFindURL.Error(), "", "text/plain; charset=utf-8"},
		},
		{
			name: "POST API 200",
			args: args{http.MethodPost, "/api/shorten", getJSONRequest()},
			want: want{http.StatusCreated, getJSONResponse(), "", "application/json; charset=UTF-8"},
		},
		{
			name: "Delete user Urls. No Content",
			args: args{http.MethodGet, "/api/user/urls", getJSONRequest()},
			want: want{http.StatusNoContent, "", "", "text/plain; charset=utf-8"},
		},
	}

	var err error
	handlers.Store, err = storage.InitDB()

	if err != nil {
		fmt.Printf("Error database init: %v\n", err)
	}

	r := CreateRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, location, contentType := testRequest(t, ts, tt.args.method, tt.args.target, tt.args.body)
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.response, body)
			assert.Equal(t, tt.want.location, location)
			assert.Equal(t, tt.want.contentType, contentType)
			bodyCloseErr := resp.Body.Close()
			if bodyCloseErr != nil {
				log.Printf("Response body close error: %v\n", bodyCloseErr)
			}
		})
	}
}

func BenchmarkHandlerSaveURL(b *testing.B) {
	w := httptest.NewRecorder()
	router := CreateRouter()

	router.HandleFunc("/", handlers.SaveShortURL)
	handlers.Store, _ = storage.InitDB()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r, _ := http.NewRequest("POST", "/", strings.NewReader(common.TestURL))
		b.StartTimer()
		router.ServeHTTP(w, r)
		result := w.Result()
		result.Body.Close()
	}
}

func BenchmarkHandlerGetURL(b *testing.B) {
	w := httptest.NewRecorder()
	router := CreateRouter()

	router.HandleFunc("/", handlers.GetURL)
	handlers.Store, _ = storage.InitDB()
	handlers.Store.Items[common.TestShortID] = common.TestURL

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r, _ := http.NewRequest("GET", "/"+common.TestShortID, nil)
		b.StartTimer()
		router.ServeHTTP(w, r)
		result := w.Result()
		result.Body.Close()
	}
}

func TestNewServer(t *testing.T) {
	type args struct {
		address string
		baseURL string
		useTLS  bool
	}
	tests := []struct {
		name    string
		args    args
		want    *server
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "OK",
			args: args{"localhost:8000", "http://localhost:8080", false},
			want: &server{
				address: "localhost:8000",
				baseURL: "http://localhost:8080",
				useTLS:  false,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewServer(tt.args.address, tt.args.baseURL, tt.args.useTLS)
			if !tt.wantErr(t, err, fmt.Sprintf("NewServer(%v, %v, %v)", tt.args.address, tt.args.baseURL, tt.args.useTLS)) {
				return
			}
			assert.Equalf(t, tt.want, got, "NewServer(%v, %v, %v)", tt.args.address, tt.args.baseURL, tt.args.useTLS)
		})
	}
}
