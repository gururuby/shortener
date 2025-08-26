/*
Package shorturl provides gRPC server implementation for URL shortening operations.

The package implements the ShortURLService gRPC service that handles:
- Creation of shortened URLs from original URLs
- Batch processing of multiple URLs in single operation
- Redirection from short aliases to original URLs
- User authentication and authorization for URL operations

Service Features:
- Single URL shortening with conflict detection
- Bulk URL processing with correlation tracking
- Secure user authentication via tokens
- Graceful error handling with appropriate gRPC status codes
- Support for both authenticated and anonymous users

Usage:
The service is designed for API clients, microservices, and internal systems
that require programmatic access to URL shortening functionality.

Protocol:
Implements the ShortURLService gRPC protocol as defined in the protobuf schema.
All methods return structured responses with error details in the response body
rather than gRPC status errors for better client compatibility.
*/
package shorturl

import (
	"context"
	"errors"

	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	shorturlUC "github.com/gururuby/shortener/internal/domain/usecase/shorturl"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	userUC "github.com/gururuby/shortener/internal/domain/usecase/user"
)

// Server implements the ShortURLService gRPC server interface.
// It handles URL shortening, batch processing, and redirection requests
// by delegating to the underlying use cases.
//
// The server provides:
// - URL shortening with duplicate detection and conflict handling
// - Batch URL processing with correlation ID tracking
// - Short URL resolution and redirection
// - Automatic user registration for anonymous requests
// - Comprehensive error handling with descriptive messages
//
// Server is thread-safe and can handle concurrent requests from multiple clients.
type Server struct {
	UnimplementedShortURLServiceServer // Embedded for forward compatibility

	// urlUC provides business logic for URL shortening operations.
	// Handles creation, batch processing, and resolution of short URLs.
	urlUC *shorturlUC.ShortURLUseCase

	// userUC provides user management operations including authentication
	// and registration. Used to associate URLs with user accounts.
	userUC *userUC.UserUseCase
}

// NewServer creates and initializes a new ShortURLService gRPC server instance.
//
// Parameters:
//   - urlUC: URL shortening use case implementation. Must not be nil and should
//     be properly configured with storage and validation dependencies.
//   - userUC: User management use case implementation. Must not be nil and should
//     handle user authentication and registration operations.
//
// Returns:
//   - *Server: Configured gRPC server instance ready to handle URL operations.
//
// Example:
//
//	urlUseCase := initializeURLUseCase()
//	userUseCase := initializeUserUseCase()
//	server := shorturl.NewServer(urlUseCase, userUseCase)
func NewServer(urlUC *shorturlUC.ShortURLUseCase, userUC *userUC.UserUseCase) *Server {
	return &Server{urlUC: urlUC, userUC: userUC}
}

// CreateShortURL handles gRPC requests to create a shortened URL from an original URL.
// It authenticates the user, validates the input, and creates the short URL.
//
// Parameters:
//   - ctx: Context for request cancellation, timeout propagation, and tracing.
//     Should include deadlines appropriate for URL creation operations.
//   - req: CreateShortURLRequest containing the authentication token and original URL.
//     AuthToken: JWT or session token for user authentication. Empty for anonymous users.
//     Url: Original URL to shorten. Must be a valid HTTP/HTTPS URL.
//
// Returns:
//   - *CreateShortURLResponse: Contains the result of the URL creation operation.
//     Result: Shortened URL string if successful.
//     Error: Descriptive error message if operation failed.
//     Conflict: Boolean indicating if the URL already exists (HTTP 409 equivalent).
//   - error: Always returns nil as errors are communicated via the response.
//
// Response Scenarios:
//   - Success: Returns Result with short URL, Error empty, Conflict false
//   - Duplicate: Returns Result with existing short URL, Error empty, Conflict true
//   - Validation Error: Returns Error with validation message, Result empty
//   - Authentication Error: Returns Error with auth failure message
//
// Example:
//
//	resp, _ := client.CreateShortURL(ctx, &pb.CreateShortURLRequest{
//	    AuthToken: "user-jwt-token",
//	    Url:       "https://example.com/long/url/path",
//	})
//	if resp.Conflict {
//	    // URL already exists, use resp.Result
//	} else if resp.Error != "" {
//	    // Handle error
//	} else {
//	    // Use new short URL: resp.Result
//	}
//
// Notes:
//   - URLs are validated according to internal validation rules
//   - Anonymous users are automatically registered with generated tokens
//   - Conflict responses include the existing short URL for client convenience
func (s *Server) CreateShortURL(ctx context.Context, req *CreateShortURLRequest) (*CreateShortURLResponse, error) {
	user, err := s.authenticateUser(ctx, req.AuthToken)
	if err != nil {
		return &CreateShortURLResponse{Error: err.Error()}, nil
	}

	shortURL, err := s.urlUC.CreateShortURL(ctx, user, req.Url)
	if err != nil {
		if errors.Is(err, ucErrors.ErrShortURLAlreadyExist) {
			return &CreateShortURLResponse{
				Result:   shortURL,
				Conflict: true,
			}, nil
		}
		return &CreateShortURLResponse{Error: err.Error()}, nil
	}

	return &CreateShortURLResponse{Result: shortURL}, nil
}

// BatchShortURLs handles gRPC requests to process multiple URLs in a single operation.
// It authenticates the user and processes all URLs in the batch atomically.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout propagation.
//     Timeouts should accommodate processing of all URLs in the batch.
//   - req: BatchShortURLsRequest containing authentication token and URL list.
//     AuthToken: JWT or session token for user authentication.
//     Urls: List of URLs to process, each with correlation ID for result matching.
//
// Returns:
//   - *BatchShortURLsResponse: Contains results for all processed URLs.
//     Urls: List of output objects with correlation IDs and short URLs.
//     Error: Descriptive error message if entire batch failed.
//
// Response Scenarios:
//   - Success: Returns Urls with results for all input URLs, Error empty
//   - Authentication Error: Returns Error with auth failure message
//   - Empty Batch: Returns Error indicating no URLs to process
//
// Example:
//
//	resp, _ := client.BatchShortURLs(ctx, &pb.BatchShortURLsRequest{
//	    AuthToken: "user-jwt-token",
//	    Urls: []*pb.BatchURLInput{
//	        {CorrelationId: "req-1", OriginalUrl: "https://example.com/1"},
//	        {CorrelationId: "req-2", OriginalUrl: "https://example.com/2"},
//	    },
//	})
//	for _, result := range resp.Urls {
//	    // Match result.CorrelationId with input
//	}
//
// Notes:
//   - Processing continues even if individual URLs in the batch fail
//   - Results maintain the same order as input URLs
//   - Correlation IDs are used to match inputs with outputs
//   - Empty batches are rejected with descriptive error
func (s *Server) BatchShortURLs(ctx context.Context, req *BatchShortURLsRequest) (*BatchShortURLsResponse, error) {
	_, err := s.authenticateUser(ctx, req.AuthToken)
	if err != nil {
		return &BatchShortURLsResponse{Error: err.Error()}, nil
	}

	if len(req.Urls) == 0 {
		return &BatchShortURLsResponse{Error: "nothing to process, empty batch"}, nil
	}

	inputs := make([]shortURLEntity.BatchShortURLInput, len(req.Urls))
	for i, url := range req.Urls {
		inputs[i] = shortURLEntity.BatchShortURLInput{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		}
	}

	outputs := s.urlUC.BatchShortURLs(ctx, inputs)
	pbOutputs := make([]*BatchURLOutput, len(outputs))
	for i, output := range outputs {
		pbOutputs[i] = &BatchURLOutput{
			CorrelationId: output.CorrelationID,
			ShortUrl:      output.ShortURL,
		}
	}

	return &BatchShortURLsResponse{Urls: pbOutputs}, nil
}

// FindShortURL handles gRPC requests to resolve a short URL alias to its original URL.
// It looks up the alias and returns the redirection target.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout propagation.
//   - req: FindShortURLRequest containing the short URL alias to resolve.
//     Alias: The short URL identifier (path segment after domain).
//
// Returns:
//   - *FindShortURLResponse: Contains the redirection target and status information.
//     Location: Original URL for redirection if found.
//     Error: Descriptive error message if alias cannot be resolved.
//     Gone: Boolean indicating the URL was permanently deleted (HTTP 410 equivalent).
//
// Response Scenarios:
//   - Success: Returns Location with original URL, Error empty, Gone false
//   - Deleted: Returns Error with deletion message, Location empty, Gone true
//   - Not Found: Returns Error with not found message, Location empty
//
// Example:
//
//	resp, _ := client.FindShortURL(ctx, &pb.FindShortURLRequest{
//	    Alias: "abc123",
//	})
//	if resp.Gone {
//	    // URL was deleted, handle appropriately
//	} else if resp.Error != "" {
//	    // Handle other errors
//	} else {
//	    // Redirect to resp.Location
//	}
//
// Notes:
//   - Does not require authentication (public endpoint)
//   - Deleted URLs return Gone status to prevent reuse
//   - Intended for use by redirection systems and clients
func (s *Server) FindShortURL(ctx context.Context, req *FindShortURLRequest) (*FindShortURLResponse, error) {
	result, err := s.urlUC.FindShortURL(ctx, req.Alias)
	if err != nil {
		if errors.Is(err, ucErrors.ErrShortURLDeleted) {
			return &FindShortURLResponse{Gone: true, Error: err.Error()}, nil
		}
		return &FindShortURLResponse{Error: err.Error()}, nil
	}

	return &FindShortURLResponse{Location: result}, nil
}

// authenticateUser handles user authentication and automatic registration.
// It authenticates existing users or registers new ones for anonymous requests.
//
// Parameters:
//   - ctx: Context for authentication operations.
//   - token: Authentication token from request. Empty string for anonymous users.
//
// Returns:
//   - *userEntity.User: Authenticated or newly registered user entity.
//   - error: Authentication or registration failure error.
//
// Notes:
//   - Internal method not exposed via gRPC interface
//   - Empty tokens trigger automatic user registration
//   - Invalid tokens fall back to new user registration
//   - Maintains user session consistency across requests
func (s *Server) authenticateUser(ctx context.Context, token string) (*userEntity.User, error) {
	if token == "" {
		return s.userUC.Register(ctx)
	}

	user, err := s.userUC.Authenticate(ctx, token)
	if err != nil {
		return s.userUC.Register(ctx)
	}
	return user, nil
}
