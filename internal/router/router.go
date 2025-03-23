package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/controllers"
	"github.com/gururuby/shortener/internal/storage"
)

func NewRouter(config *config.Config, storage storage.IStorage) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/", controllers.ShortURLCreate(config.BaseURL, storage))
	router.Get("/{alias}", controllers.ShortURLShow(storage))

	return router
}
