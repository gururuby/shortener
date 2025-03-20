package main

import (
	appConfig "github.com/gururuby/shortener/internal/infra/config"
	"github.com/gururuby/shortener/internal/infra/http_server"
)

var config = appConfig.Config{
	ServerAddress: "localhost:8080",
}

func main() {
	http_server.Run(&config)
}
