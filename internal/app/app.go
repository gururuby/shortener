package app

import (
	appConfig "github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/router"
	"github.com/gururuby/shortener/internal/storages"
	"log"
	"net/http"
)

func Run() {
	config := appConfig.NewConfig()
	storage := storages.NewMemoryStorage()

	log.Fatal(http.ListenAndServe(config.ServerAddress, router.NewRouter(config, storage)))
}
