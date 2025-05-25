package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gururuby/shortener/internal/config"
	entity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	"github.com/gururuby/shortener/internal/infra/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

type (
	headers struct {
		contentType     string
		contentEncoding string
		acceptEncoding  string
	}

	request struct {
		body      []byte
		authToken string
		headers   headers
		method    string
		path      string
	}
	compressedRequest struct {
		request
		body    *bytes.Buffer
		headers headers
		method  string
		path    string
	}
	response struct {
		headers  headers
		status   int
		location string
	}
)

func Test_App_OK(t *testing.T) {
	var (
		cfg              *config.Config
		err              error
		existingShortURL *entity.ShortURL
		user             *userEntity.User
		authToken        string
	)

	cfg, err = config.New()
	ctx := context.Background()
	require.NoError(t, err)

	app := New(cfg).Setup()
	ts := httptest.NewServer(app.Router)
	defer ts.Close()

	auth := jwt.New(cfg.Auth.SecretKey, cfg.Auth.TokenTTL)

	user, err = app.UserStorage.SaveUser(ctx)
	require.NoError(t, err)

	authToken, err = auth.SignUserID(user.ID)
	require.NoError(t, err)

	sourceURL := "https://ya.ru"
	existingShortURL, err = app.ShortURLSStorage.SaveShortURL(ctx, user, sourceURL)

	var tests = []struct {
		name     string
		request  request
		response response
		want     string
	}{
		{
			name: "when create ShortURL via http",
			request: request{
				body:    []byte(gofakeit.URL()),
				headers: headers{contentType: "text/plain; charset=utf-8"},
				method:  http.MethodPost,
				path:    "/",
			},
			response: response{
				headers: headers{contentType: "text/plain; charset=utf-8"},
				status:  http.StatusCreated,
			},
			want: `http://localhost:8080/\w{5}`,
		},
		{
			name: "when create via API",
			request: request{
				body:    []byte(fmt.Sprintf(`{"url":"%s"}`, gofakeit.URL())),
				headers: headers{contentType: "application/json"},
				method:  http.MethodPost,
				path:    "/api/shorten",
			},
			response: response{
				headers: headers{contentType: "application/json"},
				status:  http.StatusCreated,
			},
			want: `{"Result":"http://localhost:8080/\w{5}"}`,
		},
		{
			name: "when try to create via API with the same source URL",
			request: request{
				body:    []byte(`{"url":"https://ya.ru"}`),
				headers: headers{contentType: "application/json"},
				method:  http.MethodPost,
				path:    "/api/shorten",
			},
			response: response{
				headers: headers{contentType: "application/json"},
				status:  http.StatusConflict,
			},
			want: `{"Result":"http://localhost:8080/\w{5}"}`,
		},
		{
			name: "when batch creating via API",
			request: request{
				body:    []byte(fmt.Sprintf(`[{"correlation_id":"1","original_url":"%s"},{"correlation_id":"2","original_url":"%s"}]`, gofakeit.URL(), gofakeit.URL())),
				headers: headers{contentType: "application/json"},
				method:  http.MethodPost,
				path:    "/api/shorten/batch",
			},
			response: response{
				headers: headers{contentType: "application/json"},
				status:  http.StatusCreated,
			},
			want: `{"correlation_id":"1","short_url":"http://localhost:8080/\w{5}"},{"correlation_id":"2","short_url":"http://localhost:8080/\w{5}"}`,
		},
		{
			name: "when find ShortURL via http",
			request: request{
				method: http.MethodGet,
				path:   "/" + existingShortURL.Alias,
			},
			response: response{
				location: sourceURL,
				status:   http.StatusOK,
			},
		},
		{
			name: "when find user URLs via API",
			request: request{
				authToken: authToken,
				method:    http.MethodGet,
				headers:   headers{contentType: "application/json"},
				path:      "/api/user/urls",
			},
			response: response{
				headers: headers{contentType: "application/json"},
				status:  http.StatusOK,
			},
			want: `[{"short_url":"http://localhost:8080/\w{5}","original_url:"https://ya.ru"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, body := testRequest(t, ts, tt.request)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.response.status, res.StatusCode)
			if tt.response.headers.contentType != "" {
				assert.Equal(t, tt.response.headers.contentType, res.Header.Get("Content-Type"))
			}
			if tt.want != "" {
				assert.Regexp(t, regexp.MustCompile(tt.want), body)
			}

		})
	}
}

func Test_App_Compress_OK(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	app := New(cfg).Setup()

	ts := httptest.NewServer(app.Router)
	defer ts.Close()

	var tests = []struct {
		name     string
		request  compressedRequest
		response response
		want     string
	}{
		{
			name: "when send gzipped text/html",
			request: compressedRequest{
				body: zippify(t, gofakeit.URL()),
				headers: headers{
					contentType:     "text/html",
					contentEncoding: "gzip",
					acceptEncoding:  "gzip",
				},
				method: http.MethodPost,
				path:   "/",
			},
			response: response{
				headers: headers{
					contentType:     "text/plain; charset=utf-8",
					acceptEncoding:  "gzip",
					contentEncoding: "gzip",
				},
				status: http.StatusCreated,
			},
			want: `\Ahttp://localhost:8080/\w{5}\z`,
		},
		{
			name: "when content type is a application/json",
			request: compressedRequest{
				body: zippify(t, fmt.Sprintf(`{"url":"%s"}`, gofakeit.URL())),
				headers: headers{
					contentType:     "application/json",
					contentEncoding: "gzip",
					acceptEncoding:  "gzip",
				},
				method: http.MethodPost,
				path:   "/api/shorten",
			},
			response: response{
				headers: headers{
					contentType:     "application/json",
					contentEncoding: "gzip",
					acceptEncoding:  "gzip",
				},
				status: http.StatusCreated,
			},
			want: `{"Result":"http://localhost:8080/\w{5}"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var body []byte
			var zr *gzip.Reader

			resp := testCompressedRequest(t, ts, tt.request)

			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			assert.Equal(t, tt.response.status, resp.StatusCode)
			assert.Equal(t, tt.response.headers.contentType, resp.Header.Get("Content-Type"))
			assert.Equal(t, tt.response.headers.contentEncoding, resp.Header.Get("Content-Encoding"))
			assert.Equal(t, tt.response.headers.acceptEncoding, resp.Header.Get("Accept-Encoding"))

			zr, err = gzip.NewReader(resp.Body)
			require.NoError(t, err)

			body, err = io.ReadAll(zr)

			require.NoError(t, err)
			assert.Regexp(t, regexp.MustCompile(tt.want), string(body))
		})
	}
}

func Test_App_Errors(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	app := New(cfg).Setup()

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
				headers: headers{contentType: "text/plain; charset=utf-8"},
				path:    "/unknown",
				method:  http.MethodGet,
			},
			response: response{
				headers: headers{contentType: "text/plain; charset=utf-8"},
				status:  http.StatusUnprocessableEntity,
			},
			want: "record not found\n",
		},
		{
			name: "when passed incorrect url via API",
			request: request{
				body: []byte(`{"url":"//ya.ru"}`),
				headers: headers{
					contentType:     "application/json",
					contentEncoding: "application/json",
					acceptEncoding:  "application/json",
				},
				method: http.MethodPost,
				path:   "/api/shorten",
			},
			response: response{
				headers: headers{contentType: "application/json"},
				status:  http.StatusUnprocessableEntity,
			},
			want: `{"StatusCode":422,"Error":"invalid source URL, please specify valid URL"}`,
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
	var (
		err  error
		body []byte
		req  *http.Request
		resp *http.Response
	)

	req, err = http.NewRequest(r.method, ts.URL+r.path, bytes.NewReader(r.body))

	req.RequestURI = ""
	req.Header.Set("Content-Type", r.headers.contentType)
	req.Header.Set("Content-Encoding", r.headers.contentEncoding)
	req.Header.Set("Accept-Encoding", r.headers.acceptEncoding)

	if r.authToken != "" {
		authCookie := http.Cookie{Name: "Authorization", Value: r.authToken}
		req.AddCookie(&authCookie)
	}

	require.NoError(t, err)

	resp, err = ts.Client().Do(req)
	require.NoError(t, err)

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		require.NoError(t, err)
	}(resp.Body)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(body)
}

func testCompressedRequest(t *testing.T, ts *httptest.Server, r compressedRequest) *http.Response {
	req := httptest.NewRequest(r.method, ts.URL+r.path, r.body)
	req.RequestURI = ""
	req.Header.Set("Content-Type", r.headers.contentType)
	req.Header.Set("Content-Encoding", r.headers.contentEncoding)
	req.Header.Set("Accept-Encoding", r.headers.acceptEncoding)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	return resp
}

func zippify(t *testing.T, content string) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)

	_, err := zb.Write([]byte(content))
	require.NoError(t, err)

	err = zb.Close()
	require.NoError(t, err)

	return buf
}
