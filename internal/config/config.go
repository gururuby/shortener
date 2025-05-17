package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type (
	Config struct {
		App
		Auth
		Server
		Database
		FileStorage
		Log
	}

	App struct {
		AliasLength int    `env:"APP_ALIAS_LENGTH" envDefault:"5"`
		Env         string `env:"APP_ENV" envDefault:"development"`
		Name        string `env:"APP_NAME" envDefault:"Shortener"`
		Version     string `env:"APP_VERSION" envDefault:"0.0.1"`
		BaseURL     string `env:"APP_BASE_URL"`
	}

	Auth struct {
		TokenTTL  time.Duration `env:"AUTH_TOKEN_TTL" envDefault:"24h"`
		SecretKey string        `env:"AUTH_SECRET_KEY" envDefault:"secret"`
	}

	Server struct {
		Address string `env:"SERVER_ADDRESS"`
	}

	Database struct {
		ConnTryDelay time.Duration `env:"DATABASE_CONN_TRY_DELAY" envDefault:"5s"`
		ConnTryTimes int           `env:"DATABASE_CONN_TRY_TIMES" envDefault:"5"`
		Type         string        `env:"DATABASE_TYPE"`
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
	var err error

	err = godotenv.Load(".env")
	if err != nil {
		log.Print("Error loading .env file")
	}

	if err = env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	flag.Parse()

	if cfg.Database.DSN == "" {
		if cfg.FileStorage.Path == "" {
			cfg.Database.Type = "memory"
		} else {
			cfg.Database.Type = "file"
		}
	} else {
		cfg.Database.Type = "postgresql"
	}

	return &cfg, nil
}

func (c *Config) AppInfo() string {
	return fmt.Sprintf("%s v%s (%s)", c.App.Name, c.App.Version, c.App.Env)
}

func init() {
	flag.StringVar(&cfg.Server.Address, "a", "localhost:8080", "Server address")
	flag.StringVar(&cfg.App.BaseURL, "b", "http://localhost:8080", "Base URL of short URLs")
	flag.StringVar(&cfg.Database.DSN, "d", "", "URL to database")
	flag.StringVar(&cfg.FileStorage.Path, "f", "/tmp/db.json", "ShortURLs storage file")
}
