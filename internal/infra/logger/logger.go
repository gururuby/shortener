/*
Package logger provides centralized logging functionality for the application.

It features:
- Thread-safe singleton logger initialization
- Environment-specific logging configurations
- Configurable log levels
- Structured logging via zap logger
- Production and development logging presets
*/
package logger

import (
	"go.uber.org/zap"
	"log"
	"sync"
)

// Log is the global logger instance that should be used throughout the application.
// It is initialized by calling Setup() and provides structured logging methods.
var Log *zap.Logger

// Setup initializes the global logger with the specified environment and log level.
// This function is safe for concurrent use and will only initialize the logger once.
//
// Parameters:
//   - appENV: Application environment ("production" or any other value for development)
//   - logLevel: Desired log level ("debug", "info", "warn", "error")
//
// Note: If initialization fails, the function will log the error and exit the program.
func Setup(appENV, logLevel string) {
	var initLogger sync.Once

	initLogger.Do(func() {
		var cfg zap.Config
		var err error

		switch appENV {
		case "production":
			cfg = zap.NewProductionConfig()
		default:
			cfg = zap.NewDevelopmentConfig()
		}

		cfg.Level = buildLogLevel(logLevel)

		if Log, err = cfg.Build(); err != nil {
			log.Fatalf("cannot init logger: %s", err)
		}
	})
}

// buildLogLevel converts a string log level to zap's AtomicLevel.
// This is an internal helper function used during logger setup.
//
// Parameters:
//   - logLevel: String representation of log level
//
// Returns:
//   - zap.AtomicLevel: Configured log level (defaults to InfoLevel for invalid inputs)
func buildLogLevel(logLevel string) zap.AtomicLevel {
	var lvl zap.AtomicLevel

	switch logLevel {
	case "debug":
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		lvl = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		lvl = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return lvl
}
