package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gururuby/shortener/internal/config"
	entity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	"github.com/gururuby/shortener/internal/infra/jwt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
)

func Example() {
	var (
		cfg              *config.Config
		existingShortURL *entity.ShortURL
		user             *userEntity.User
		authToken        string
	)

	cfg, _ = config.New()
	ctx := context.Background()

	app := New(cfg).Setup()
	ts := httptest.NewServer(app.Router)
	defer ts.Close()

	auth := jwt.New(cfg.Auth.SecretKey, cfg.Auth.TokenTTL)

	user, _ = app.UserStorage.SaveUser(ctx)

	authToken, _ = auth.SignUserID(user.ID)

	sourceURL := "https://ya.ru"
	existingShortURL, _ = app.ShortURLSStorage.SaveShortURL(ctx, user, sourceURL)
	urls := []string{
		gofakeit.URL(),
		gofakeit.URL(),
		gofakeit.URL(),
		gofakeit.URL(),
	}

	var examples = []struct {
		name    string
		match   string
		request request
	}{
		{
			name: "when create ShortURL via http",
			request: request{
				body:    []byte(urls[0]),
				headers: headers{contentType: "text/plain; charset=utf-8"},
				method:  http.MethodPost,
				path:    "/",
			},
			match: `http://localhost:8080/\w{5}`,
		},
		{
			name: "when create via API",
			request: request{
				body:    []byte(fmt.Sprintf(`{"url":"%s"}`, urls[1])),
				headers: headers{contentType: "application/json"},
				method:  http.MethodPost,
				path:    "/api/shorten",
			},
			match: `{"Result":"http://localhost:8080/\w{5}"}`,
		},
		{
			name: "when try to create via API with the same source URL",
			request: request{
				body:    []byte(`{"url":"https://ya.ru"}`),
				headers: headers{contentType: "application/json"},
				method:  http.MethodPost,
				path:    "/api/shorten",
			},
			match: `{"Result":"http://localhost:8080/\w{5}"}`,
		},
		{
			name: "when batch creating via API",
			request: request{
				body:    []byte(fmt.Sprintf(`[{"correlation_id":"1","original_url":"%s"},{"correlation_id":"2","original_url":"%s"}]`, urls[2], urls[3])),
				headers: headers{contentType: "application/json"},
				method:  http.MethodPost,
				path:    "/api/shorten/batch",
			},
			match: `{"correlation_id":"1","short_url":"http://localhost:8080/\w{5}"},{"correlation_id":"2","short_url":"http://localhost:8080/\w{5}"}`,
		},
		{
			name: "when find ShortURL via http",
			request: request{
				method: http.MethodGet,
				path:   "/" + existingShortURL.Alias,
			},
			match: `<!doctype html>`,
		},
		{
			name: "when find user URLs via API",
			request: request{
				authToken: authToken,
				method:    http.MethodGet,
				headers:   headers{contentType: "application/json"},
				path:      "/api/user/urls",
			},
			match: `[{"short_url":"http://localhost:8080/\w{5}","original_url:"https://ya.ru"}]`,
		},
	}

	for _, ex := range examples {
		res := testExampleRequest(ts, ex.request)
		reg := regexp.MustCompile(ex.match)
		fmt.Println("Response matched with:", ex.match, reg.MatchString(res))
	}

	// Output:
	// Response matched with: http://localhost:8080/\w{5} true
	// Response matched with: {"Result":"http://localhost:8080/\w{5}"} true
	// Response matched with: {"Result":"http://localhost:8080/\w{5}"} true
	// Response matched with: {"correlation_id":"1","short_url":"http://localhost:8080/\w{5}"},{"correlation_id":"2","short_url":"http://localhost:8080/\w{5}"} true
	// Response matched with: <!doctype html> true
	// Response matched with: [{"short_url":"http://localhost:8080/\w{5}","original_url:"https://ya.ru"}] true
}

func testExampleRequest(ts *httptest.Server, r request) string {
	var (
		err  error
		body []byte
		req  *http.Request
		resp *http.Response
	)

	req, err = http.NewRequest(r.method, ts.URL+r.path, bytes.NewReader(r.body))
	if err != nil {
		panic(err)
	}

	req.RequestURI = ""
	req.Header.Set("Content-Type", r.headers.contentType)
	req.Header.Set("Content-Encoding", r.headers.contentEncoding)
	req.Header.Set("Accept-Encoding", r.headers.acceptEncoding)

	if r.authToken != "" {
		authCookie := http.Cookie{Name: "Authorization", Value: r.authToken}
		req.AddCookie(&authCookie)
	}

	resp, err = ts.Client().Do(req)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	body, err = io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	return string(body)
}
