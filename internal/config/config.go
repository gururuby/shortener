/*
Package config handles application configuration management including:
- Environment variable parsing
- Command-line flag processing
- Configuration defaults
- Configuration validation

The package supports configuration from multiple sources:
1. .env files
2. Environment variables
3. Command-line flags
4. Default values

Configuration is organized into logical sections (App, Auth, Server, etc.)
for better maintainability.
*/
package config

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Config represents the complete application configuration.
// It aggregates all configuration subsections.
type Config struct {
	Server
	FileStorage
	Log
	App
	Auth
	Database
}

// App contains application metadata and general settings.
type App struct {
	Env         string `env:"APP_ENV" envDefault:"development"`
	Name        string `env:"APP_NAME" envDefault:"Shortener"`
	Version     string `env:"APP_VERSION" envDefault:"0.0.1"`
	BaseURL     string `env:"APP_BASE_URL"`
	AliasLength int    `env:"APP_ALIAS_LENGTH" envDefault:"5"`
}

// Auth contains JWT authentication settings.
type Auth struct {
	SecretKey string        `env:"AUTH_SECRET_KEY" envDefault:"secret"`
	TokenTTL  time.Duration `env:"AUTH_TOKEN_TTL" envDefault:"24h"`
}

// HTTPS contains HTTPS server configuration
type HTTPS struct {
	Enabled  bool   `env:"ENABLE_HTTPS" envDefault:"false"`
	CertFile string `env:"HTTPS_CERT_FILE"`
	KeyFile  string `env:"HTTPS_KEY_FILE"`
}

// Server contains HTTP server configuration.
type Server struct {
	Address string `env:"SERVER_ADDRESS"` // Server listen address (host:port)
	HTTPS
}

// Database contains database connection settings.
type Database struct {
	Type         string        `env:"DATABASE_TYPE"`
	DSN          string        `env:"DATABASE_DSN"`
	ConnTryDelay time.Duration `env:"DATABASE_CONN_TRY_DELAY" envDefault:"5s"`
	ConnTryTimes int           `env:"DATABASE_CONN_TRY_TIMES" envDefault:"5"`
}

// FileStorage contains settings for file-based storage.
type FileStorage struct {
	Path string `env:"FILE_STORAGE_PATH"` // Path to storage file
}

// Log contains logging configuration.
type Log struct {
	Level string `env:"LOG_LEVEL" envDefault:"info"` // Logging level (debug/info/warn/error)
}

var cfg Config // Global configuration instance

// New loads and initializes application configuration from multiple sources:
// 1. .env file (if present)
// 2. Environment variables
// 3. Command-line flags
//
// Returns:
// - *Config: Loaded configuration
// - error: Any error that occurred during loading
func New() (*Config, error) {
	var err error

	// Try loading .env file (ignore if not found)
	err = godotenv.Load(".env")
	if err != nil {
		log.Print("Error loading .env file")
	}

	// Parse environment variables
	if err = env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config error: %v", err)
	}

	// Parse command-line flags
	flag.Parse()

	// Determine storage type based on provided configuration
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

// AppInfo generates a formatted string with application information.
// Format: "<Name> v<Version> (<Env>)"
func (c *Config) AppInfo() string {
	return fmt.Sprintf("%s v%s (%s)", c.App.Name, c.App.Version, c.App.Env)
}

// init registers command-line flags with their default values.
func init() {
	flag.StringVar(&cfg.Server.Address, "a", "localhost:8080", "Server address (host:port)")
	flag.StringVar(&cfg.App.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&cfg.Database.DSN, "d", "", "Database connection string (DSN)")
	flag.StringVar(&cfg.FileStorage.Path, "f", "/tmp/db.json", "Path to file storage")
	flag.BoolVar(&cfg.Server.HTTPS.Enabled, "s", true, "Run HTTPS server")
}
