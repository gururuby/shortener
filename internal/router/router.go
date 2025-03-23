package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/controllers"
	"github.com/gururuby/shortener/internal/storage"
)

func Router(config *config.Config, storage storage.StorageInterface) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/", controllers.ShortURLCreate(config.PublicAddress, storage))
	router.Get("/{alias}", controllers.ShortURLShow(storage))

	return router
}
