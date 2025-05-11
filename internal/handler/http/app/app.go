//go:generate mockgen -destination=./mocks/mock.go -package=mocks . AppUseCase

package handler

import (
	"context"
	"fmt"
	"net/http"
)

const (
	pingDBPath = "/ping"
)

type Router interface {
	Get(path string, h http.HandlerFunc)
}

type AppUseCase interface {
	PingDB(ctx context.Context) error
}

type handler struct {
	uc     AppUseCase
	router Router
}

func Register(router Router, uc AppUseCase) {
	h := handler{router: router, uc: uc}
	h.router.Get(pingDBPath, h.PingDB())
}

func (h *handler) PingDB() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		if r.Method != http.MethodGet {
			http.Error(w, fmt.Sprintf("HTTP method %s is not allowed", r.Method), http.StatusMethodNotAllowed)
			return
		}

		err = h.uc.PingDB(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
