package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

type (
	Config struct {
		App    App
		Server Server
		DB     DB
	}

	App struct {
		Env     string `env:"APP_ENV" envDefault:"development"`
		Name    string `env:"APP_NAME" envDefault:"Shortener"`
		Version string `env:"APP_VERSION" envDefault:"0.0.1"`
		BaseURL string `env:"APP_BASE_URL"`
	}

	Server struct {
		Address string `env:"SERVER_ADDRESS"`
	}

	DB struct {
		Type string `env:"DB_TYPE" envDefault:"memory"`
	}
)

var cfg Config

func New() (*Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	flag.Parse()

	return &cfg, nil
}

func (c *Config) AppInfo() string {
	return fmt.Sprintf("%s v%s (%s)", c.App.Name, c.App.Version, c.App.Env)
}

func init() {
	flag.StringVar(&cfg.Server.Address, "a", "localhost:8080", "Server address")
	flag.StringVar(&cfg.App.BaseURL, "b", "http://localhost:8080", "Base URL of short URLs")
}
