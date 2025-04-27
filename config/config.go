package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"time"
)

type (
	Config struct {
		App
		Server
		Database
		FileStorage
		Log
	}

	App struct {
		AliasLength           int    `env:"APP_ALIAS_LENGTH" envDefault:"5"`
		Env                   string `env:"APP_ENV" envDefault:"development"`
		MaxGenerationAttempts int    `env:"APP_MAX_GENERATION_ATTEMPTS" envDefault:"5"`
		Name                  string `env:"APP_NAME" envDefault:"Shortener"`
		Version               string `env:"APP_VERSION" envDefault:"0.0.1"`
		BaseURL               string `env:"APP_BASE_URL"`
	}

	Server struct {
		Address string `env:"SERVER_ADDRESS"`
	}

	Database struct {
		ConnTryDelay time.Duration `env:"DATABASE_CONN_TRY_DELAY" envDefault:"5s"`
		ConnTryTimes int           `env:"DATABASE_CONN_TRY_TIMES" envDefault:"5"`
		Type         string        `env:"DATABASE_TYPE" envDefault:"postgresql"`
		DSN          string        `env:"DATABASE_DSN"`
	}

	FileStorage struct {
		Path string `env:"FILE_STORAGE_PATH"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL" envDefault:"info"`
	}
)

var cfg Config

func New() (*Config, error) {
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
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
	flag.StringVar(&cfg.Database.DSN, "d", "postgresql://postgres:pass@0.0.0.0:5432/shortener?sslmode=disable", "URL to database")
	flag.StringVar(&cfg.FileStorage.Path, "f", "/tmp/db.json", "ShortURLs storage file")
}
