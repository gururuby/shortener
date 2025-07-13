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
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"log"
	"time"
)

// Config represents the complete application configuration.
// It aggregates all configuration subsections.
type Config struct {
	App         // Application metadata and settings
	Auth        // Authentication settings
	Server      // HTTP server settings
	Database    // Database connection settings
	FileStorage // File storage settings
	Log         // Logging configuration
}

// App contains application metadata and general settings.
type App struct {
	AliasLength int    `env:"APP_ALIAS_LENGTH" envDefault:"5"`  // Length of generated URL aliases
	Env         string `env:"APP_ENV" envDefault:"development"` // Runtime environment (development/production/etc.)
	Name        string `env:"APP_NAME" envDefault:"Shortener"`  // Application name
	Version     string `env:"APP_VERSION" envDefault:"0.0.1"`   // Application version
	BaseURL     string `env:"APP_BASE_URL"`                     // Base URL for shortened links
}

// Auth contains JWT authentication settings.
type Auth struct {
	TokenTTL  time.Duration `env:"AUTH_TOKEN_TTL" envDefault:"24h"`     // JWT token time-to-live
	SecretKey string        `env:"AUTH_SECRET_KEY" envDefault:"secret"` // JWT signing key
}

// Server contains HTTP server configuration.
type Server struct {
	Address string `env:"SERVER_ADDRESS"` // Server listen address (host:port)
}

// Database contains database connection settings.
type Database struct {
	ConnTryDelay time.Duration `env:"DATABASE_CONN_TRY_DELAY" envDefault:"5s"` // Delay between connection attempts
	ConnTryTimes int           `env:"DATABASE_CONN_TRY_TIMES" envDefault:"5"`  // Number of connection attempts
	Type         string        `env:"DATABASE_TYPE"`                           // Database type (memory/file/postgresql)
	DSN          string        `env:"DATABASE_DSN"`                            // Database connection string
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
}
