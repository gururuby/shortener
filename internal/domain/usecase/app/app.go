//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Storage

package usecase

import (
	"context"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/app/errors"
)

type Storage interface {
	IsDBReady(ctx context.Context) error
}

type AppUseCase struct {
	storage Storage
}

func NewAppUseCase(storage Storage) *AppUseCase {
	return &AppUseCase{
		storage: storage,
	}
}

func (uc *AppUseCase) PingDB(ctx context.Context) error {
	if err := uc.storage.IsDBReady(ctx); err != nil {
		return ucErrors.ErrAppDBIsNotReady
	}
	return nil
}
