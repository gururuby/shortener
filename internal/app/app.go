package app

import (
	appConfig "github.com/gururuby/shortener/internal/app/config"
	"github.com/gururuby/shortener/internal/app/router"
	"log"
	"net/http"
)

var config = appConfig.NewConfig()

func Run() {
	log.Fatal(http.ListenAndServe(config.ServerAddress, router.Router()))
}
