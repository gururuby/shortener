/*
Package stats provides gRPC server implementation for application statistics and metrics.

The package implements the StatsService gRPC service that provides:
- Real-time application metrics and usage statistics
- System health and capacity monitoring data
- Aggregate counts of URLs and users for dashboarding
- Operational intelligence for system administrators
*/
package stats

import (
	"context"

	statsUC "github.com/gururuby/shortener/internal/domain/usecase/stats"
)

// Server implements the StatsService gRPC server interface.
// It handles statistics retrieval requests by delegating to the underlying use case.

// Server is thread-safe and optimized for high read throughput with minimal
// blocking operations. Suitable for frequent polling by monitoring systems.
type Server struct {
	UnimplementedStatsServiceServer // Embedded for forward compatibility

	// useCase provides business logic for statistics aggregation and retrieval.
	// Handles data collection, counting, and metric calculation operations.
	// Must be configured with appropriate data access layers for accurate statistics.
	useCase *statsUC.StatsUseCase
}

// NewServer creates and initializes a new StatsService gRPC server instance.
//
// Parameters:
//   - useCase: Statistics use case implementation. Must not be nil and should
//     be properly configured with data access dependencies for accurate counting.
//     The use case should provide efficient counting operations for large datasets.
//
// Returns:
//   - *Server: Configured gRPC server instance ready to handle statistics requests.
func NewServer(useCase *statsUC.StatsUseCase) *Server {
	return &Server{useCase: useCase}
}

// GetStats handles gRPC requests to retrieve current application statistics.
// It returns aggregate counts of URLs and users in the system.
//
// Parameters:
//   - ctx: Context for request cancellation, timeout propagation, and tracing.
//     Timeouts should be set appropriately for statistics collection operations
//     which may involve database queries or distributed counting.
//   - req: GetStatsRequest containing optional parameters for statistics retrieval.
//     Currently empty but reserved for future filtering or sampling options.
//
// Returns:
//   - *GetStatsResponse: Contains the current application statistics and status.
//     UrlsCount: Total number of shortened URLs in the system.
//     UsersCount: Total number of registered users in the system.
//     Error: Descriptive error message if statistics retrieval failed.
//   - error: Always returns nil as errors are communicated via the response.
func (s *Server) GetStats(ctx context.Context, req *GetStatsRequest) (*GetStatsResponse, error) {
	stats, err := s.useCase.GetStats(ctx)
	if err != nil {
		return &GetStatsResponse{Error: err.Error()}, nil
	}

	return &GetStatsResponse{
		UrlsCount:  int64(stats.URLsCount),
		UsersCount: int64(stats.UsersCount),
	}, nil
}
