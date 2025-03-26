package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/handlers"
	"net/http"
)

type Storage interface {
	Save(string, string) (string, bool)
	Find(string) (string, bool)
}

type Handler interface {
	Create() http.HandlerFunc
	Show() http.HandlerFunc
}

func NewRouter(config *config.Config, storage Storage) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	urlsHandler := handlers.NewURLsHandler(config, storage)

	router.Post("/", urlsHandler.Create())
	router.Get("/{alias}", urlsHandler.Show())

	return router
}
