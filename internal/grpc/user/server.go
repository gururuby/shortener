// Package user provides gRPC server implementation for user operations.
package user

import (
	"context"

	userUC "github.com/gururuby/shortener/internal/domain/usecase/user"
)

// Server implements UserService gRPC server.
type Server struct {
	UnimplementedUserServiceServer
	useCase *userUC.UserUseCase
}

// NewServer creates a new UserService server instance.
func NewServer(useCase *userUC.UserUseCase) *Server {
	return &Server{useCase: useCase}
}

// GetURLs retrieves all shortened URLs belonging to the authenticated user.
// Automatically registers new users if authentication fails.
func (s *Server) GetURLs(ctx context.Context, req *GetURLsRequest) (*GetURLsResponse, error) {
	user, err := s.useCase.Authenticate(ctx, req.AuthToken)
	if err != nil {
		// Try to register new user if authentication fails
		user, err = s.useCase.Register(ctx)
		if err != nil {
			return &GetURLsResponse{Error: err.Error()}, nil
		}
	}

	urls, err := s.useCase.GetURLs(ctx, user)
	if err != nil {
		return &GetURLsResponse{Error: err.Error()}, nil
	}

	pbURLs := make([]*UserURL, len(urls))
	for i, url := range urls {
		pbURLs[i] = &UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		}
	}

	return &GetURLsResponse{Urls: pbURLs}, nil
}

// DeleteURLs removes the specified short URLs belonging to the authenticated user.
// Requires non-empty list of aliases to delete.
func (s *Server) DeleteURLs(ctx context.Context, req *DeleteURLsRequest) (*DeleteURLsResponse, error) {
	user, err := s.useCase.Authenticate(ctx, req.AuthToken)
	if err != nil {
		return &DeleteURLsResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	if len(req.Aliases) == 0 {
		return &DeleteURLsResponse{
			Success: false,
			Error:   "no aliases passed to delete short urls",
		}, nil
	}

	s.useCase.DeleteURLs(ctx, user, req.Aliases)
	return &DeleteURLsResponse{Success: true}, nil
}
