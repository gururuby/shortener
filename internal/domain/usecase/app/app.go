//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Storage

/*
Package usecase implements the application's business logic layer.

It contains:
- Core business rules and workflows
- Use case implementations
- Error handling specific to business operations
*/
package usecase

import (
	"context"

	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/app/errors"
)

// Storage defines the interface for storage operations required by the application use cases.
type Storage interface {
	// IsDBReady checks if the database connection is healthy.
	// Returns:
	// - error: If database is not ready or connection fails
	IsDBReady(ctx context.Context) error
}

// AppUseCase implements application-level use cases.
// It coordinates between the application and storage layers.
type AppUseCase struct {
	storage Storage // Storage layer interface
}

// NewAppUseCase creates a new instance of AppUseCase.
// Parameters:
// - storage: Implementation of the Storage interface
// Returns:
// - *AppUseCase: Initialized application use case instance
func NewAppUseCase(storage Storage) *AppUseCase {
	return &AppUseCase{
		storage: storage,
	}
}

// PingDB checks the database connection status.
// Parameters:
// - ctx: Context for cancellation and timeouts
// Returns:
// - error: Returns ErrAppDBIsNotReady if database is unavailable, nil otherwise
func (uc *AppUseCase) PingDB(ctx context.Context) error {
	if err := uc.storage.IsDBReady(ctx); err != nil {
		return ucErrors.ErrAppDBIsNotReady
	}
	return nil
}
