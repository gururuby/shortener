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
5. JSON configuration files

Configuration is organized into logical sections (App, Auth, Server, etc.)
for better maintainability.
*/
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

// Config represents the complete application configuration.
// It aggregates all configuration subsections including server settings,
// authentication parameters, database configuration and logging setup.
type Config struct {
	Server      Server      // HTTP/HTTPS server configuration
	FileStorage FileStorage // File storage settings
	Log         Log         // Logging configuration
	App         App         // Application metadata
	Auth        Auth        // Authentication settings
	Database    Database    // Database connection parameters
}

// App contains application metadata and general settings.
type App struct {
	Env             string        `env:"APP_ENV" envDefault:"development"`      // Application environment (development/production)
	Name            string        `env:"APP_NAME" envDefault:"Shortener"`       // Application name
	Version         string        `env:"APP_VERSION" envDefault:"0.0.1"`        // Application version
	BaseURL         string        `env:"APP_BASE_URL"`                          // Base URL for generated links
	AliasLength     int           `env:"APP_ALIAS_LENGTH" envDefault:"5"`       // Default length for generated aliases
	ShutdownTimeout time.Duration `env:"APP_SHUTDOWN_TIMEOUT" envDefault:"30s"` // Graceful shutdown timeout
}

// Auth contains JWT authentication settings.
type Auth struct {
	SecretKey string        `env:"AUTH_SECRET_KEY" envDefault:"secret"` // Secret key for JWT tokens
	TokenTTL  time.Duration `env:"AUTH_TOKEN_TTL" envDefault:"24h"`     // Token time-to-live duration
}

// HTTPS contains HTTPS server configuration.
type HTTPS struct {
	Enabled  bool   `env:"ENABLE_HTTPS" envDefault:"false"` // Enable HTTPS server
	CertFile string `env:"HTTPS_CERT_FILE"`                 // Path to SSL certificate file
	KeyFile  string `env:"HTTPS_KEY_FILE"`                  // Path to SSL private key file
}

// Server contains HTTP server configuration.
type Server struct {
	Address       string        `env:"SERVER_ADDRESS"`                           // Server listen address (host:port)
	ReadTimeout   time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"5s"`      // Maximum duration for reading request
	TrustedSubnet string        `env:"TRUSTED_SUBNET" envDefault:"127.0.0.1/24"` //Subnet setting for restrict access to specific endpoints
	WriteTimeout  time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10s"`    // Maximum duration for writing response
	IdleTimeout   time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"120s"`    // Maximum idle connection duration
	HTTPS         HTTPS         // HTTPS-specific configuration
	GRPC          GRPC          // GRPC-specific configuration
}

// GRPC contains GRPC server configuration
type GRPC struct {
	Enabled               bool          `env:"GRPC_ENABLED" envDefault:"false"`
	ConnectionTimeout     time.Duration `env:"GRPC_CONNECTION_TIMEOUT" envDefault:"120s"`
	Address               string        `env:"GRPC_ADDRESS" envDefault:":50051"`
	MaxConnectionIdle     time.Duration `env:"GRPC_MAX_CONNECTION_IDLE" envDefault:"2h"`
	MaxConnectionAge      time.Duration `env:"GRPC_MAX_CONNECTION_AGE" envDefault:"30m"`
	MaxConnectionAgeGrace time.Duration `env:"GRPC_MAX_CONNECTION_AGE_GRACE" envDefault:"5m"`
	KeepaliveTime         time.Duration `env:"GRPC_KEEPALIVE_TIME" envDefault:"2h"`
	KeepaliveTimeout      time.Duration `env:"GRPC_KEEPALIVE_TIMEOUT" envDefault:"20s"`
	MinKeepaliveTime      time.Duration `env:"GRPC_MIN_KEEPALIVE_TIME" envDefault:"10s"`
	PermitWithoutStream   bool          `env:"GRPC_PERMIT_WITHOUT_STREAM" envDefault:"true"`
}

// Database contains database connection settings.
type Database struct {
	Type         string        `env:"DATABASE_TYPE"`                           // Database type (postgresql/mysql/file/memory)
	DSN          string        `env:"DATABASE_DSN"`                            // Data Source Name (connection string)
	ConnTryDelay time.Duration `env:"DATABASE_CONN_TRY_DELAY" envDefault:"5s"` // Delay between connection attempts
	ConnTryTimes int           `env:"DATABASE_CONN_TRY_TIMES" envDefault:"5"`  // Number of connection attempts
}

// FileStorage contains settings for file-based storage.
type FileStorage struct {
	Path string `env:"FILE_STORAGE_PATH"` // Path to storage file
}

// Log contains logging configuration.
type Log struct {
	Level string `env:"LOG_LEVEL" envDefault:"info"` // Logging level (debug/info/warn/error)
}

var (
	cfg         Config // Global configuration instance
	jsonCfgName string // Name of JSON config file
)

// New loads and initializes application configuration from multiple sources:
// 1. .env file (if present)
// 2. Environment variables
// 3. Command-line flags
// 4. JSON configuration file (if specified)
//
// The loading order follows the priority:
// 1. Command-line flags (highest priority)
// 2. Environment variables
// 3. .env file
// 4. JSON config file
// 5. Default values (lowest priority)
//
// Returns:
// - *Config: Loaded configuration
// - error: Any error that occurred during loading
func New() (*Config, error) {
	var err error

	// Load from JSON config file if specified
	if jsonCfgName != "" {
		err = loadConfigFromJSON(jsonCfgName, cfg)
		if err != nil {
			log.Printf("Error loading config from %s file", jsonCfgName)
		}
	}

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

// loadConfigFromJSON reads and parses JSON configuration file into Config struct.
// The function expects the path to a valid JSON file matching the Config structure.
// Returns error if file cannot be read or contains invalid configuration.
func loadConfigFromJSON(path string, cfg Config) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(file, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// AppInfo generates a formatted string with application information.
// The format is: "<Name> v<Version> (<Env>)"
// Example: "Shortener v1.0.0 (production)"
func (c *Config) AppInfo() string {
	return fmt.Sprintf("%s v%s (%s)", c.App.Name, c.App.Version, c.App.Env)
}

// init registers command-line flags with their default values.
// The flags are registered when the package is initialized.
func init() {
	flag.StringVar(&cfg.Server.Address, "a", "localhost:8080", "Server address (host:port)")
	flag.StringVar(&cfg.App.BaseURL, "b", "http://localhost:8080", "Base URL for shortened links")
	flag.StringVar(&jsonCfgName, "c", "", "Name of config file")
	flag.StringVar(&cfg.Database.DSN, "d", "", "Database connection string (DSN)")
	flag.StringVar(&cfg.FileStorage.Path, "f", "/tmp/db.json", "Path to file storage")
	flag.StringVar(&cfg.Server.TrustedSubnet, "t", "", "Trusted subnet")
	flag.BoolVar(&cfg.Server.HTTPS.Enabled, "s", true, "Run HTTPS server")

}
