//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DAO

package usecase

import (
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
)

type DAO interface {
	IsDBReady() error
}

type AppUseCase struct {
	dao DAO
}

func NewAppUseCase(dao DAO) *AppUseCase {
	return &AppUseCase{
		dao: dao,
	}
}

func (uc *AppUseCase) PingDB() error {
	if err := uc.dao.IsDBReady(); err != nil {
		return ucErrors.ErrAppDBIsNotReady
	}
	return nil
}
