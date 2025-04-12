package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

type request struct {
	body        io.Reader
	contentType string
	method      string
	path        string
}

type response struct {
	contentType string
	status      int
	location    string
}

func TestAppOkRequests(t *testing.T) {
	app := Setup()
	ts := httptest.NewServer(app.Router)
	defer ts.Close()

	sourceURL := "https://example.com"
	existingAlias, _ := app.Storage.Save(sourceURL)

	var tests = []struct {
		name     string
		request  request
		response response
		want     string
	}{
		{
			name: "when create ShortURL via http",
			request: request{
				body:        strings.NewReader("https://ya.ru"),
				contentType: "text/plain; charset=utf-8",
				method:      http.MethodPost,
				path:        "/",
			},
			response: response{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusCreated,
			},
			want: "^http://localhost:8080/\\w{5}$",
		},
		{
			name: "when create via API",
			request: request{
				body:        bytes.NewBuffer([]byte(`{"url":"https://ya.ru"}`)),
				contentType: "application/json",
				method:      http.MethodPost,
				path:        "/api/shorten",
			},
			response: response{
				contentType: "application/json",
				status:      http.StatusCreated,
			},
			want: "\\{\"Result\":\"http://localhost:8080/\\w{5}\"\\}",
		},
		{
			name: "when find ShortURL via http",
			request: request{
				method: http.MethodGet,
				path:   "/" + existingAlias,
			},
			response: response{
				location: sourceURL,
				status:   http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, body := testRequest(t, ts, tt.request)
			err := res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.response.status, res.StatusCode)
			if tt.response.contentType != "" {
				assert.Equal(t, tt.response.contentType, res.Header.Get("Content-Type"))
			}
			if tt.want != "" {
				assert.Regexp(t, regexp.MustCompile(tt.want), body)
			}

		})
	}
}

func TestAppErrorRequests(t *testing.T) {
	app := Setup()
	ts := httptest.NewServer(app.Router)
	defer ts.Close()

	var tests = []struct {
		name     string
		request  request
		response response
		want     string
	}{
		{
			name: "when cannot find ShortURL",
			request: request{
				body:        nil,
				contentType: "text/plain; charset=utf-8",
				path:        "/unknown",
				method:      http.MethodGet,
			},
			response: response{
				contentType: "text/plain; charset=utf-8",
				status:      http.StatusUnprocessableEntity,
			},
			want: "source URL not found\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, result := testRequest(t, ts, tt.request)
			err := res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.response.status, res.StatusCode)
			if tt.want != "" {
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, r request) (*http.Response, string) {
	req, err := http.NewRequest(r.method, ts.URL+r.path, r.body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", r.contentType)

	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, string(respBody)
}
