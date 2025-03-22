package app

import (
	appConfig "github.com/gururuby/shortener/internal/app/config"
	"github.com/gururuby/shortener/internal/app/controllers"
	"github.com/gururuby/shortener/internal/app/repos"
	"log"
	"net/http"
)

var config = appConfig.NewConfig()

func Run() {
	mux := http.NewServeMux()

	storage := repos.NewShortURLsRepo()

	mux.HandleFunc("/", controllers.ShortURLCreate(storage))
	mux.HandleFunc("/{id}", controllers.ShortURLShow(storage))

	log.Fatal(http.ListenAndServe(config.ServerAddress, mux))
}
