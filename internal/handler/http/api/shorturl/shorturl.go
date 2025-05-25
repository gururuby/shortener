//go:generate mockgen -destination=./mocks/mock.go -package=mocks . ShortURLUseCase,UserUseCase

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	apiErrors "github.com/gururuby/shortener/internal/handler/http/api/shorturl/errors"
	"net/http"
	"time"
)

const (
	authCookieName        = "Authorization"
	createShortURLTimeout = time.Second * 30
	createShortURLPath    = "/api/shorten"

	batchShortURLsTimeout = time.Second * 60
	batchShortURLsPath    = "/api/shorten/batch"
)

type Router interface {
	Post(path string, h http.HandlerFunc)
}

type ShortURLUseCase interface {
	CreateShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (string, error)
	FindShortURL(ctx context.Context, alias string) (string, error)
	BatchShortURLs(ctx context.Context, urls []shortURLEntity.BatchShortURLInput) []shortURLEntity.BatchShortURLOutput
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

type errorResponse struct {
	StatusCode int
	Error      string
}

type (
	createShortURLDTO struct {
		request struct {
			URL string
		}
		response struct {
			Result string
		}
	}

	batchShortURLsDTO struct {
		inputURLs  []shortURLEntity.BatchShortURLInput
		outputURLs []shortURLEntity.BatchShortURLOutput
	}
)

func Register(router Router, userUC UserUseCase, urlUC ShortURLUseCase) {
	h := handler{router: router, userUC: userUC, urlUC: urlUC}
	h.router.Post(batchShortURLsPath, h.BatchShortURLs())
	h.router.Post(createShortURLPath, h.CreateShortURL())
}

func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
			user       *userEntity.User
			statusCode = http.StatusCreated
			shortURL   string
			response   []byte
			dto        createShortURLDTO
			errRes     errorResponse
		)

		ctx, cancel := context.WithTimeout(r.Context(), createShortURLTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
			returnErrResponse(errRes, w)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(&dto.request); err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusBadRequest
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

		shortURL, err = h.urlUC.CreateShortURL(ctx, user, dto.request.URL)

		if err != nil {
			if errors.Is(err, ucErrors.ErrShortURLAlreadyExist) {
				statusCode = http.StatusConflict
			} else {
				errRes.Error = err.Error()
				errRes.StatusCode = http.StatusUnprocessableEntity
				returnErrResponse(errRes, w)
				return
			}

		}

		dto.response.Result = shortURL
		response, err = json.Marshal(dto.response)

		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusInternalServerError
			returnErrResponse(errRes, w)
			return
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

func (h *handler) BatchShortURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err      error
			response []byte
			dto      batchShortURLsDTO
			errRes   errorResponse
		)

		ctx, cancel := context.WithTimeout(r.Context(), batchShortURLsTimeout)
		defer cancel()

		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			errRes.Error = fmt.Sprintf("HTTP method %s is not allowed", r.Method)
			errRes.StatusCode = http.StatusMethodNotAllowed
			returnErrResponse(errRes, w)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(&dto.inputURLs); err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		if len(dto.inputURLs) == 0 {
			errRes.Error = apiErrors.ErrAPIEmptyBatch.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		dto.outputURLs = h.urlUC.BatchShortURLs(ctx, dto.inputURLs)
		response, err = json.Marshal(dto.outputURLs)

		if err != nil {
			errRes.Error = err.Error()
			errRes.StatusCode = http.StatusInternalServerError
			returnErrResponse(errRes, w)
			return
		}

		w.WriteHeader(http.StatusCreated)

		if _, err = w.Write(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
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
