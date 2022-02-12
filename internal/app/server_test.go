package app

import (
	"github.com/fd239/go_url_shortener/internal/app/_const"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string, string) {
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

	defer resp.Body.Close()

	location := resp.Header.Get("location")

	return resp, string(respBody), location
}

func TestRouter(t *testing.T) {
	type want struct {
		code     int
		response string
		location string
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
			args: args{http.MethodPost, "/", strings.NewReader(_const.TestUrl)},
			want: want{http.StatusCreated, _const.TestShortIdFullUrl, ""},
		},
		{
			name: "POST 400 Empty body",
			args: args{http.MethodPost, "/", nil},
			want: want{http.StatusBadRequest, _const.ErrMsg_EmptyBody, ""},
		},
		{
			name: "GET 307",
			args: args{http.MethodGet, "/" + _const.TestShortId, nil},
			want: want{http.StatusTemporaryRedirect, "", _const.TestUrl},
		},
		{
			name: "GET 405 No ID in request",
			args: args{http.MethodGet, "/", nil},
			want: want{http.StatusMethodNotAllowed, "", ""},
		},
		{
			name: "GET 400 No URL in map",
			args: args{http.MethodGet, "/123", nil},
			want: want{http.StatusBadRequest, _const.ErrMsg_NoUrlInMap, ""},
		},
	}

	r := CreateRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body, location := testRequest(t, ts, tt.args.method, tt.args.target, tt.args.body)
			body = strings.TrimSuffix(body, "\n")
			assert.Equal(t, resp.StatusCode, tt.want.code)
			assert.Equal(t, body, tt.want.response)
			assert.Equal(t, location, tt.want.location)
		})
	}
}
