package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/gururuby/shortener/internal/middleware"
	"net/http"
)

type Router interface {
	Post(path string, h http.HandlerFunc)
	Get(path string, h http.HandlerFunc)
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}

func Setup() Router {
	router := chi.NewRouter()
	router.Use(middleware.Logging)
	router.Use(middleware.Compression)

	return router
}
