/*
Package server provides HTTP server implementation with:
- Configurable HTTP/HTTPS support
- Graceful shutdown handling
- Proper timeout management
- Signal handling for termination
*/
package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/server/errors"
	"go.uber.org/zap"
)

// Router defines the interface for HTTP request routing.
// Implementations should handle HTTP requests and route them to appropriate handlers.
type Router interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// DB defines the interface for database shutdown operations.
// Implementations should provide graceful shutdown capability for database connections.
type DB interface {
	// Shutdown gracefully closes database connections with the given context.
	// Returns error if shutdown fails or context expires.
	Shutdown(context.Context) error
}

// Server represents an HTTP server with graceful shutdown capabilities.
// It manages the server lifecycle including startup, shutdown and error handling.
type Server struct {
	config  *config.Config // Application configuration including server settings
	router  Router         // HTTP request router implementation
	backend *http.Server   // Underlying HTTP server instance
	db      DB             // Database interface for graceful shutdown
}

// New creates and configures a new Server instance.
// Parameters:
//   - router: HTTP request router implementation
//   - cfg: Application configuration containing server settings
//   - db: Database instance that supports graceful shutdown
//
// Returns:
//   - *Server: Configured server instance ready to run
func New(router Router, cfg *config.Config, db DB) *Server {
	return &Server{
		router:  router,
		config:  cfg,
		backend: createHTTPServer(router, cfg),
		db:      db,
	}
}

// Run starts the HTTP/HTTPS server and blocks until shutdown.
// It handles:
//   - Server startup in HTTP or HTTPS mode based on configuration
//   - Graceful shutdown on receiving termination signals
//   - Error handling and logging
func (s *Server) Run() {
	serverErr := make(chan error, 1) // Channel for server startup errors

	go func() {
		if s.config.Server.HTTPS.Enabled {
			serverErr <- s.startHTTPS()
		} else {
			serverErr <- s.startHTTP()
		}
	}()

	s.waitForShutdown(serverErr) // Wait for shutdown signal or error
}

// createHTTPServer initializes the http.Server with configured timeouts.
// Parameters:
//   - router: HTTP request router
//   - cfg: Configuration containing timeout settings
//
// Returns:
//   - *http.Server: Configured HTTP server instance
func createHTTPServer(router Router, cfg *config.Config) *http.Server {
	return &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
}

// startHTTPS starts the server in HTTPS mode with TLS encryption.
// Returns:
//   - error: If server fails to start or TLS configuration is invalid
func (s *Server) startHTTPS() error {
	if err := validateTLSConfig(s.config); err != nil {
		return err
	}

	logger.Log.Info("HTTPS server starting",
		zap.String("certFile", s.config.Server.HTTPS.CertFile),
		zap.String("keyFile", s.config.Server.HTTPS.KeyFile),
	)
	return s.backend.ListenAndServeTLS(
		s.config.Server.HTTPS.CertFile,
		s.config.Server.HTTPS.KeyFile,
	)
}

// startHTTP starts the server in HTTP mode without encryption.
// Returns:
//   - error: If server fails to start
func (s *Server) startHTTP() error {
	logger.Log.Info("HTTP server starting")
	return s.backend.ListenAndServe()
}

// validateTLSConfig verifies HTTPS configuration is valid.
// Parameters:
//   - cfg: Configuration containing TLS settings
//
// Returns:
//   - error: If certificate or key files are not specified
func validateTLSConfig(cfg *config.Config) error {
	if cfg.Server.HTTPS.CertFile == "" || cfg.Server.HTTPS.KeyFile == "" {
		logger.Log.Error("Invalid TLS configuration",
			zap.String("certFile", cfg.Server.HTTPS.CertFile),
			zap.String("keyFile", cfg.Server.HTTPS.KeyFile),
		)
		return errors.ErrServerInvalidTLSConfig
	}
	return nil
}

// waitForShutdown listens for server errors or termination signals.
// Parameters:
//   - serverErr: Channel receiving server startup/run errors
func (s *Server) waitForShutdown(serverErr <-chan error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case err := <-serverErr:
		logger.Log.Error("Server runtime error", zap.Error(err))
	case sig := <-interrupt:
		s.handleGracefulShutdown(sig)
	}
}

// handleGracefulShutdown performs graceful server shutdown.
// Parameters:
//   - sig: Received termination signal
func (s *Server) handleGracefulShutdown(sig os.Signal) {
	logger.Log.Info("Initiating graceful shutdown",
		zap.String("signal", sig.String()),
		zap.Duration("timeout", s.config.App.ShutdownTimeout),
	)

	ctx, cancel := context.WithTimeout(context.Background(), s.config.App.ShutdownTimeout)
	defer cancel()

	// Shutdown HTTP server
	if err := s.backend.Shutdown(ctx); err != nil {
		logger.Log.Error("Graceful shutdown failed, forcing exit", zap.Error(err))
		s.forceShutdown()
	}

	// Shutdown database
	if err := s.db.Shutdown(ctx); err != nil {
		logger.Log.Error("DB Graceful shutdown failed", zap.Error(err))
	}

	logger.Log.Info("Server shutdown completed")
}

// forceShutdown immediately terminates all server connections.
// Used as fallback when graceful shutdown fails.
func (s *Server) forceShutdown() {
	if err := s.backend.Close(); err != nil {
		logger.Log.Error("Forced shutdown error", zap.Error(err))
	}
}
