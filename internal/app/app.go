package app

import (
	"flag"
	"github.com/caarlos0/env/v6"
	appConfig "github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/repos"
	"github.com/gururuby/shortener/internal/router"
	"log"
	"net/http"
)

func Run() {
	config := appConfig.NewConfig()
	// Override config from ENVs
	err := env.Parse(config)
	if err != nil {
		log.Fatal(err)
	}

	// Override config via flags
	flag.StringVar(&config.ServerAddress, "a", "", "Server address")
	flag.StringVar(&config.BaseURL, "b", "", "Base URL of short URLs")

	flag.Parse()

	// Setup default config values
	if config.ServerAddress == "" {
		config.ServerAddress = "localhost:8080"
	}

	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}

	storage := repos.NewShortURLsRepo()

	log.Fatal(http.ListenAndServe(config.ServerAddress, router.NewRouter(config, storage)))
}
