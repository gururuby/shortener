/*
Package server provides gRPC server implementation with:
- Configurable gRPC server support
- Graceful shutdown handling
- Proper timeout management
- Signal handling for termination
- TLS support for secure connections
*/
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gururuby/shortener/internal/config"
	"github.com/gururuby/shortener/internal/infra/logger"
	grpcErrors "github.com/gururuby/shortener/internal/infra/server/grpc/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// GRPCServer defines the interface for gRPC service registration.
// Implementations should register gRPC services with the server.
type GRPCServer interface {
	// RegisterServices registers all gRPC services with the server
	RegisterServices(*grpc.Server)
}

// DB defines the interface for database shutdown operations.
// Implementations should provide graceful shutdown capability for database connections.
type DB interface {
	// Shutdown gracefully closes database connections with the given context.
	// Returns error if shutdown fails or context expires.
	Shutdown(context.Context) error
}

// Server represents a gRPC server with graceful shutdown capabilities.
// It manages the server lifecycle including startup, shutdown and error handling.
type Server struct {
	config   *config.Config // Application configuration including server settings
	grpcImpl GRPCServer     // gRPC service implementation
	backend  *grpc.Server   // Underlying gRPC server instance
	db       DB             // Database interface for graceful shutdown
	listener net.Listener   // Network listener
}

// New creates and configures a new gRPC Server instance.
// Parameters:
//   - grpcImpl: gRPC service implementation
//   - cfg: Application configuration containing server settings
//   - db: Database instance that supports graceful shutdown
//
// Returns:
//   - *Server: Configured server instance ready to run
//   - error: If server configuration fails
func New(grpcImpl GRPCServer, cfg *config.Config, db DB) (*Server, error) {
	s := &Server{
		config:   cfg,
		grpcImpl: grpcImpl,
		db:       db,
	}

	// Create gRPC server with options
	opts, err := s.createGRPCOptions()
	if err != nil {
		return nil, err
	}

	s.backend = grpc.NewServer(opts...)

	// Register services
	grpcImpl.RegisterServices(s.backend)

	// Create listener
	listener, err := net.Listen("tcp", cfg.Server.GRPC.Address)
	if err != nil {
		return nil, err
	}
	s.listener = listener

	return s, nil
}

// Run starts the gRPC server and blocks until shutdown.
// It handles:
//   - Server startup in secure or insecure mode based on configuration
//   - Graceful shutdown on receiving termination signals
//   - Error handling and logging
func (s *Server) Run() {
	serverErr := make(chan error, 1) // Channel for server startup errors

	go func() {
		logger.Log.Info("gRPC server starting",
			zap.String("address", s.config.Server.GRPC.Address),
			zap.Bool("tls", s.config.Server.HTTPS.Enabled),
		)
		serverErr <- s.backend.Serve(s.listener)
	}()

	s.waitForShutdown(serverErr) // Wait for shutdown signal or error
}

// createGRPCOptions creates gRPC server options based on configuration.
// Returns:
//   - []grpc.ServerOption: Configured gRPC server options
//   - error: If TLS configuration fails
func (s *Server) createGRPCOptions() ([]grpc.ServerOption, error) {
	var opts []grpc.ServerOption

	// Add TLS credentials if enabled
	if s.config.Server.HTTPS.Enabled {
		creds, err := s.createTLSCredentials()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Add keepalive options
	opts = append(opts, s.createKeepaliveOptions()...)

	// Add timeout options
	opts = append(opts, s.createTimeoutOptions()...)

	return opts, nil
}

// createTLSCredentials creates TLS credentials for secure gRPC connections.
// Returns:
//   - credentials.TransportCredentials: Configured TLS credentials
//   - error: If TLS configuration is invalid
func (s *Server) createTLSCredentials() (credentials.TransportCredentials, error) {
	if err := validateTLSConfig(s.config); err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(
		s.config.Server.HTTPS.CertFile,
		s.config.Server.HTTPS.KeyFile,
	)
	if err != nil {
		logger.Log.Error("Failed to load TLS certificates",
			zap.String("certFile", s.config.Server.HTTPS.CertFile),
			zap.String("keyFile", s.config.Server.HTTPS.KeyFile),
			zap.Error(err),
		)
		return nil, grpcErrors.ErrGRPCServerInvalidTLSConfig
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	return credentials.NewTLS(cfg), nil
}

// createKeepaliveOptions creates keepalive options for gRPC server.
// Returns:
//   - []grpc.ServerOption: Keepalive server options
func (s *Server) createKeepaliveOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     s.config.Server.GRPC.MaxConnectionIdle,
			MaxConnectionAge:      s.config.Server.GRPC.MaxConnectionAge,
			MaxConnectionAgeGrace: s.config.Server.GRPC.MaxConnectionAgeGrace,
			Time:                  s.config.Server.GRPC.KeepaliveTime,
			Timeout:               s.config.Server.GRPC.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             s.config.Server.GRPC.MinKeepaliveTime,
			PermitWithoutStream: s.config.Server.GRPC.PermitWithoutStream,
		}),
	}
}

// createTimeoutOptions creates timeout options for gRPC server.
// Returns:
//   - []grpc.ServerOption: Timeout server options
func (s *Server) createTimeoutOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ConnectionTimeout(s.config.Server.GRPC.ConnectionTimeout),
	}
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
		return grpcErrors.ErrGRPCServerInvalidTLSConfig
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
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Log.Error("gRPC server runtime error", zap.Error(err))
		}
	case sig := <-interrupt:
		s.handleGracefulShutdown(sig)
	}
}

// handleGracefulShutdown performs graceful server shutdown.
// Parameters:
//   - sig: Received termination signal
func (s *Server) handleGracefulShutdown(sig os.Signal) {
	logger.Log.Info("Initiating graceful gRPC shutdown",
		zap.String("signal", sig.String()),
		zap.Duration("timeout", s.config.App.ShutdownTimeout),
	)

	ctx, cancel := context.WithTimeout(context.Background(), s.config.App.ShutdownTimeout)
	defer cancel()

	// Graceful gRPC server shutdown
	s.backend.GracefulStop()

	// Shutdown database
	if err := s.db.Shutdown(ctx); err != nil {
		logger.Log.Error("DB Graceful shutdown failed", zap.Error(err))
	}

	logger.Log.Info("gRPC server shutdown completed")
}

// GetBackend returns the underlying gRPC server instance.
// Useful for testing and advanced server management.
func (s *Server) GetBackend() *grpc.Server {
	return s.backend
}

// Stop immediately stops the gRPC server.
// Used for testing and emergency shutdown scenarios.
func (s *Server) Stop() {
	s.backend.Stop()
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return
		}
	}
}
