package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	shortensPath = "/api/shorten"
)

type Router interface {
	Post(path string, h http.HandlerFunc)
}

type UseCase interface {
	CreateShortURL(sourceURL string) (string, error)
	FindShortURL(alias string) (string, error)
}

type handler struct {
	uc     UseCase
	router Router
}

type createShortURLDTO struct {
	request struct {
		URL string
	}
	response struct {
		Result string
	}
}

func Register(router Router, uc UseCase) {
	h := handler{router: router, uc: uc}
	h.router.Post(shortensPath, h.CreateShortURL())
}

func (h *handler) CreateShortURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var shortURL string
		var response []byte
		var dto createShortURLDTO

		if r.Method != http.MethodPost {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		if err = json.NewDecoder(r.Body).Decode(&dto.request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		shortURL, err = h.uc.CreateShortURL(dto.request.URL)
		dto.response.Result = shortURL

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		response, err = json.Marshal(dto.response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		if _, err = w.Write(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
