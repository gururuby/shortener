package app

import (
	appConfig "github.com/gururuby/shortener/internal/app/config"
	"github.com/gururuby/shortener/internal/app/controllers"
	"log"
	"net/http"
)

var config = appConfig.NewConfig()

func Run() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", controllers.ShortURLCreate)
	mux.HandleFunc("/{id}", controllers.ShortURLShow)

	log.Fatal(http.ListenAndServe(config.ServerAddress, mux))
}
