package app

import (
	appConfig "github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/repos"
	"github.com/gururuby/shortener/internal/router"
	"log"
	"net/http"
)

func Run() {
	config := appConfig.NewConfig()
	storage := repos.NewShortURLsRepo()

	log.Fatal(http.ListenAndServe(config.ServerAddress, router.NewRouter(config, storage)))
}
