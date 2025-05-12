package main

import (
	"github.com/gururuby/shortener/internal/app"
	"github.com/gururuby/shortener/internal/config"
	"log"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("cannot setup config: %s", err)
	}
	app.New(cfg).Setup().Run()
}
