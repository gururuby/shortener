package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gururuby/shortener/internal/app/controllers"
	"github.com/gururuby/shortener/internal/app/repos"
)

var storage = repos.NewShortURLsRepo()

func Router() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Post("/", controllers.ShortURLCreate(storage))
	router.Get("/{alias}", controllers.ShortURLShow(storage))

	return router
}
