package http_server

import (
	"github.com/gururuby/shortener/internal/app/controllers/short_urls_controller"
	"github.com/gururuby/shortener/internal/infra/config"
	"log"
	"net/http"
)

func Run(config *config.Config) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", short_urls_controller.Create)
	mux.HandleFunc("/{id}", short_urls_controller.Show)

	log.Fatal(http.ListenAndServe(config.ServerAddress, mux))
}
