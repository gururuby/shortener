package app

import (
	"flag"
	appConfig "github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/repos"
	"github.com/gururuby/shortener/internal/router"
	"log"
	"net/http"
)

func Run() {
	config := new(appConfig.Config)
	storage := repos.NewShortURLsRepo()
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Base address of running server")
	flag.StringVar(&config.PublicAddress, "b", "localhost:8080", "Base address of short links")

	flag.Parse()

	log.Fatal(http.ListenAndServe(config.ServerAddress, router.Router(config, storage)))
}
