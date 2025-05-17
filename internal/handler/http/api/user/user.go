//go:generate mockgen -destination=./mocks/mock.go -package=mocks . UserUseCase

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	"github.com/gururuby/shortener/internal/domain/usecase/user"
	"net/http"
	"time"
)

const (
	authCookieName     = "Authorization"
	getUserURLsTimeout = time.Second * 30
	getUserURLsPath    = "/api/user/urls"
)

type Router interface {
	Get(path string, h http.HandlerFunc)
}

type UserUseCase interface {
	GetURLs(ctx context.Context, user *userEntity.User) ([]*usecase.UserShortURL, error)
	Authenticate(ctx context.Context, token string) (*userEntity.User, error)
	Register(ctx context.Context) (*userEntity.User, error)
}

type handler struct {
	userUC UserUseCase
	router Router
}

type errorResponse struct {
	StatusCode int
	Error      string
}

func Register(router Router, userUC UserUseCase) {
	h := handler{router: router, userUC: userUC}
	h.router.Get(getUserURLsPath, h.GetUserURLs())
}

func (h *handler) GetUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			statusCode int
			response   []byte
			errRes     errorResponse
			user       *userEntity.User
			userURLs   []*usecase.UserShortURL
		)

		ctx, cancel := context.WithTimeout(r.Context(), getUserURLsTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodGet {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
			returnErrResponse(errRes, w)
			return
		}

		user, err = h.authUser(ctx, r, w)
		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusUnprocessableEntity
			returnErrResponse(errRes, w)
			return
		}

		userURLs, err = h.userUC.GetURLs(ctx, user)
		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusInternalServerError
			returnErrResponse(errRes, w)
			return
		}

		if len(userURLs) == 0 {
			statusCode = http.StatusNoContent
			response = []byte("{}")
		} else {
			statusCode = http.StatusOK
			response, err = json.Marshal(userURLs)
			if err != nil {
				errRes.Error = err.Error()
				errRes.StatusCode = http.StatusInternalServerError
				returnErrResponse(errRes, w)
				return
			}
		}

		w.WriteHeader(statusCode)

		if _, err = w.Write(response); err != nil {
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

func returnErrResponse(errResp errorResponse, w http.ResponseWriter) {
	w.WriteHeader(errResp.StatusCode)
	response, err := json.Marshal(errResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if _, err = w.Write(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
