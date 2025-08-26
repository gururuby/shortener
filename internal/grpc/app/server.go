/*
Package app provides gRPC server implementation for application health checks and monitoring.

The package contains:
- gRPC service definitions for application health monitoring
- Server implementation bridging gRPC requests to application use cases
- Error handling and response formatting for gRPC clients

Service Features:
- Database connectivity testing via PingDB endpoint
- Health status reporting
- Graceful error handling and status propagation

Usage:
Typically used by monitoring systems, load balancers, and infrastructure tools
to verify application and database health status.
*/
package app

import (
	"context"

	appUC "github.com/gururuby/shortener/internal/domain/usecase/app"
)

// Server implements the gRPC AppService server interface.
// It handles incoming gRPC requests and delegates to the underlying use case.
//
// The server provides:
// - Health check endpoints for application monitoring
// - Database connectivity verification
// - Standardized error responses for gRPC clients
//
// Server is safe for concurrent use by multiple goroutines.
type Server struct {
	UnimplementedAppServiceServer // Embedded for forward compatibility

	// useCase provides application-level business logic operations.
	// Handles database connectivity checks and health status verification.
	useCase *appUC.AppUseCase
}

// NewServer creates and initializes a new AppService gRPC server instance.
//
// Parameters:
//   - useCase: Application use case implementation that provides business logic
//     for health checks and database operations. Must not be nil.
//
// Returns:
//   - *Server: Configured gRPC server instance ready to handle requests.
//
// Example:
//
//	appUseCase := initializeAppUseCase()
//	server := app.NewServer(appUseCase)
func NewServer(useCase *appUC.AppUseCase) *Server {
	return &Server{useCase: useCase}
}

// PingDB handles gRPC requests to check database connectivity status.
// It verifies the database connection and returns the operational status.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout propagation.
//     The context should include deadlines and tracing information.
//   - req: PingDBRequest containing optional parameters for the health check.
//     Currently empty but reserved for future extensibility.
//
// Returns:
//   - *PingDBResponse: Contains success status and optional error information.
//     Success: true indicates successful database connection.
//     Error: contains descriptive message if connection failed.
//   - error: Always returns nil as errors are communicated via the response
//     to maintain gRPC status code compatibility.
//
// Response Codes:
//   - Success: Returns response with Success=true and no error
//   - Failure: Returns response with Success=false and error description
//
// Example gRPC client usage:
//
//	resp, err := client.PingDB(ctx, &pb.PingDBRequest{})
//	if resp.Success {
//	    // Database is healthy
//	} else {
//	    // Handle error: resp.Error contains details
//	}
//
// Notes:
//   - The method is idempotent and safe for frequent polling
//   - Response time depends on database network latency and configuration
//   - Errors are logged internally for monitoring purposes
func (s *Server) PingDB(ctx context.Context, req *PingDBRequest) (*PingDBResponse, error) {
	err := s.useCase.PingDB(ctx)
	if err != nil {
		return &PingDBResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &PingDBResponse{Success: true}, nil
}
