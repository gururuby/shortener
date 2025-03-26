package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

var config Config

func NewConfig() *Config {
	err := env.Parse(&config)
	if err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	return &config
}

func init() {
	flag.StringVar(&config.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Base URL of short URLs")
}
