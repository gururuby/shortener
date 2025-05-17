package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gururuby/shortener/internal/domain/entity"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	handlerErrors "github.com/gururuby/shortener/internal/handler/errors"
	"net/http"
	"time"
)

const (
	createShortURLTimeout = time.Second * 30
	createShortURLPath    = "/api/shorten"

	batchShortURLsTimeout = time.Second * 60
	batchShortURLsPath    = "/api/shorten/batch"
)

type Router interface {
	Post(path string, h http.HandlerFunc)
}

type ShortURLUseCase interface {
	CreateShortURL(ctx context.Context, sourceURL string) (string, error)
	BatchShortURLs(ctx context.Context, urls []entity.BatchShortURLInput) []entity.BatchShortURLOutput
}

type handler struct {
	uc     ShortURLUseCase
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
		inputURLs  []entity.BatchShortURLInput
		outputURLs []entity.BatchShortURLOutput
	}
)

func Register(router Router, uc ShortURLUseCase) {
	h := handler{router: router, uc: uc}
	h.router.Post(batchShortURLsPath, h.BatchShortURLs())
	h.router.Post(createShortURLPath, h.CreateShortURL())
}

func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			err        error
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

		shortURL, err = h.uc.CreateShortURL(ctx, dto.request.URL)

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
			errRes.Error = handlerErrors.ErrAPIEmptyBatch.Error()
			errRes.StatusCode = http.StatusBadRequest
			returnErrResponse(errRes, w)
			return
		}

		dto.outputURLs = h.uc.BatchShortURLs(ctx, dto.inputURLs)
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
