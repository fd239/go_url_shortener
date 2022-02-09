package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenerHandler(t *testing.T) {
	type want struct {
		code     int
		response string
		location string
	}
	type args struct {
		method string
		target string
		body   io.Reader
		users  map[string]string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "POST 200",
			args: args{http.MethodPost, "/", strings.NewReader("http://cjdr17afeihmk.biz/kdni9/z9womotrbk"), map[string]string{}},
			want: want{http.StatusCreated, "http://localhost:8080/a7a40cddf446bc419af5737fc92f1757", ""},
		},
		{
			name: "POST 400 Empty body",
			args: args{http.MethodPost, "/", nil, map[string]string{}},
			want: want{http.StatusBadRequest, "Empty body", ""},
		},
		{
			name: "GET 307",
			args: args{http.MethodGet, "/a7a40cddf446bc419af5737fc92f1757", nil, map[string]string{"a7a40cddf446bc419af5737fc92f1757": "http://localhost:8080/a7a40cddf446bc419af5737fc92f1757"}},
			want: want{http.StatusTemporaryRedirect, "", "http://localhost:8080/a7a40cddf446bc419af5737fc92f1757"},
		},
		{
			name: "GET 400 No ID in request",
			args: args{http.MethodGet, "/", nil, map[string]string{}},
			want: want{http.StatusBadRequest, "No ID in request", ""},
		},
		{
			name: "GET 400 No URL in map",
			args: args{http.MethodGet, "/123", nil, map[string]string{}},
			want: want{http.StatusBadRequest, "No URL in map", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.args.method, tt.args.target, tt.args.body)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(ShortenerHandler(tt.args.users))
			h.ServeHTTP(w, request)
			res := w.Result()

			if res.StatusCode != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			defer res.Body.Close()
			resBody, _ := io.ReadAll(res.Body)
			resString := string(resBody)
			trimmedString := strings.TrimSuffix(resString, "\n")

			if trimmedString != tt.want.response {
				t.Errorf("Expected body %s, got %s", tt.want.response, string(resBody))
			}

			if res.Header.Get("location") != tt.want.location {
				t.Errorf("Expected location %s, got %s", tt.want.location, res.Header.Get(tt.want.location))
			}

		})
	}
}
