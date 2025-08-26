/*
Package registry provides centralized service registration for gRPC servers.

The package serves as a coordination point for registering multiple gRPC services
onto a single gRPC server instance. It implements the server.GRPCServer interface
to provide a unified way to initialize all application services.

Key Features:
- Centralized service registration management
- Implementation of GRPCServer interface for compatibility
- Dependency injection for all gRPC service implementations
- Simplified server initialization with multiple services

Usage:
Primarily used during server startup to register all available gRPC services
in a single operation. This package acts as the composition root for gRPC services.

Example:

	appServer := appGRPC.NewServer(appUseCase)
	statsServer := statsGRPC.NewServer(statsUseCase)
	userServer := userGRPC.NewServer(userUseCase)
	shorturlServer := shortURLGRPC.NewServer(shortURLUseCase, userUseCase)

	registry := NewServiceRegistry(appServer, statsServer, userServer, shorturlServer)
	grpcServer := grpc.NewServer()
	registry.RegisterServices(grpcServer)
*/
package registry

import (
	appGRPC "github.com/gururuby/shortener/internal/grpc/app"
	shortURLGRPC "github.com/gururuby/shortener/internal/grpc/shorturl"
	statsGRPC "github.com/gururuby/shortener/internal/grpc/stats"
	userGRPC "github.com/gururuby/shortener/internal/grpc/user"
	"google.golang.org/grpc"
)

// ServiceRegistry implements server.GRPCServer interface and manages
// the registration of all gRPC services for the application.
//
// The registry acts as a facade that coordinates the initialization of
// multiple gRPC services, providing a single point of registration for
// the main gRPC server instance.
//
// ServiceRegistry is immutable after creation and safe for concurrent use.
type ServiceRegistry struct {
	// appServer handles application health checks and database connectivity
	// monitoring via the AppService gRPC service.
	appServer *appGRPC.Server

	// statsServer provides application statistics and metrics reporting
	// via the StatsService gRPC service.
	statsServer *statsGRPC.Server

	// userServer manages user-related operations including URL management
	// and authentication via the UserService gRPC service.
	userServer *userGRPC.Server

	// shorturlServer handles URL shortening operations, redirection,
	// and batch processing via the ShortURLService gRPC service.
	shorturlServer *shortURLGRPC.Server
}

// NewServiceRegistry creates a new ServiceRegistry instance with all
// required gRPC service implementations.
//
// Parameters:
//   - appServer: AppService server implementation for health checks.
//     Must not be nil.
//   - statsServer: StatsService server implementation for metrics.
//     Must not be nil.
//   - userServer: UserService server implementation for user operations.
//     Must not be nil.
//   - shorturlServer: ShortURLService server implementation for URL shortening.
//     Must not be nil.
//
// Returns:
//   - *ServiceRegistry: Configured service registry ready to register services.
//
// Example:
//
//	registry := NewServiceRegistry(
//	    appServer,
//	    statsServer,
//	    userServer,
//	    shorturlServer,
//	)
//
// Panics:
//
//	The function does not panic but individual services may have their own
//	validation requirements that could cause panics during registration.
func NewServiceRegistry(
	appServer *appGRPC.Server,
	statsServer *statsGRPC.Server,
	userServer *userGRPC.Server,
	shorturlServer *shortURLGRPC.Server,
) *ServiceRegistry {
	return &ServiceRegistry{
		appServer:      appServer,
		statsServer:    statsServer,
		userServer:     userServer,
		shorturlServer: shorturlServer,
	}
}

// RegisterServices registers all managed gRPC services with the provided
// gRPC server instance. This method implements the server.GRPCServer interface.
//
// Parameters:
//   - grpcServer: The target gRPC server instance where services will be registered.
//     Must not be nil and should be properly configured before registration.
//
// The method registers the following services in order:
//  1. AppService - for health checks and monitoring
//  2. StatsService - for application statistics
//  3. UserService - for user management operations
//  4. ShortURLService - for URL shortening functionality
//
// Usage:
//
//	grpcServer := grpc.NewServer()
//	registry.RegisterServices(grpcServer)
//	// Now grpcServer has all services registered and ready to serve
//
// Notes:
//   - This method should be called only once per gRPC server instance
//   - The order of registration does not affect service functionality
//   - All services are registered regardless of individual service state
//   - The method does not validate service dependencies or readiness
func (s *ServiceRegistry) RegisterServices(grpcServer *grpc.Server) {
	// Register all gRPC services
	appGRPC.RegisterAppServiceServer(grpcServer, s.appServer)
	statsGRPC.RegisterStatsServiceServer(grpcServer, s.statsServer)
	userGRPC.RegisterUserServiceServer(grpcServer, s.userServer)
	shortURLGRPC.RegisterShortURLServiceServer(grpcServer, s.shorturlServer)
}
