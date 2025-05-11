//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DAO

package usecase

import (
	"context"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
)

type DAO interface {
	IsDBReady(ctx context.Context) error
}

type AppUseCase struct {
	dao DAO
}

func NewAppUseCase(dao DAO) *AppUseCase {
	return &AppUseCase{
		dao: dao,
	}
}

func (uc *AppUseCase) PingDB(ctx context.Context) error {
	if err := uc.dao.IsDBReady(ctx); err != nil {
		return ucErrors.ErrAppDBIsNotReady
	}
	return nil
}
