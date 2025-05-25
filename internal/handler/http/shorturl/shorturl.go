//go:generate mockgen -destination=./mocks/mock.go -package=mocks . UserUseCase,ShortURLUseCase

package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	"io"
	"net/http"
	"time"
)

const (
	authCookieName        = "Authorization"
	createShortURLTimeout = time.Second * 30
	shortensPath          = "/"
	shortenPath           = "/{alias}"
)

type Router interface {
	Post(path string, h http.HandlerFunc)
	Get(path string, h http.HandlerFunc)
}

type ShortURLUseCase interface {
	CreateShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (string, error)
	FindShortURL(ctx context.Context, alias string) (string, error)
	BatchShortURLs(ctx context.Context, urls []entity.BatchShortURLInput) []entity.BatchShortURLOutput
}

type UserUseCase interface {
	Authenticate(ctx context.Context, token string) (*userEntity.User, error)
	Register(ctx context.Context) (*userEntity.User, error)
}

type handler struct {
	userUC UserUseCase
	urlUC  ShortURLUseCase
	router Router
}

func Register(router Router, urlUC ShortURLUseCase, userUC UserUseCase) {
	h := handler{router: router, urlUC: urlUC, userUC: userUC}
	h.router.Get(shortenPath, h.FindShortURL())
	h.router.Post(shortensPath, h.CreateShortURL())

}

func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			user       *userEntity.User
			reqBody    []byte
			shortURL   string
			statusCode = http.StatusCreated
		)

		ctx, cancel := context.WithTimeout(r.Context(), createShortURLTimeout)
		defer cancel()

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		reqBody, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sourceURL := string(reqBody)

		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}(r.Body)

		user, err = h.authUser(ctx, r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		shortURL, err = h.urlUC.CreateShortURL(r.Context(), user, sourceURL)

		if err != nil {
			if errors.Is(err, ucErrors.ErrShortURLAlreadyExist) {
				statusCode = http.StatusConflict
			} else {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(statusCode)

		_, err = io.WriteString(w, shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *handler) authUser(ctx context.Context, r *http.Request, w http.ResponseWriter) (*userEntity.User, error) {
	var (
		authCookie *http.Cookie
		user       *userEntity.User
		err        error
	)

	authCookie, err = r.Cookie(authCookieName)
	// If auth cookie was not passed
	if err != nil && errors.Is(err, http.ErrNoCookie) {
		// Register new User
		if user, err = h.userUC.Register(ctx); err != nil {
			return nil, err
		}

	} else { // If auth cookie exist, try to authenticate User
		if user, err = h.userUC.Authenticate(ctx, authCookie.Value); err != nil {
			// If auth cookie is invalid or user not found try to register new user
			if user, err = h.userUC.Register(ctx); err != nil {
				return nil, err
			}
		}
	}
	// Setup auth cookie
	http.SetCookie(w, &http.Cookie{Name: authCookieName, Value: user.AuthToken})

	return user, nil
}

func (h *handler) FindShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}
		result, err := h.urlUC.FindShortURL(r.Context(), r.URL.Path)

		if err != nil {
			if errors.Is(err, ucErrors.ErrShortURLDeleted) {
				http.Error(w, err.Error(), http.StatusGone)
				return
			}
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Location", result)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
