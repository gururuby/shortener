//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Storage

package usecase

import (
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/app/errors"
)

type Storage interface {
	IsDBReady() error
}

type AppUseCase struct {
	storage Storage
}

func NewAppUseCase(storage Storage) *AppUseCase {
	return &AppUseCase{
		storage: storage,
	}
}

func (uc *AppUseCase) PingDB() error {
	if err := uc.storage.IsDBReady(); err != nil {
		return ucErrors.ErrAppDBIsNotReady
	}
	return nil
}
