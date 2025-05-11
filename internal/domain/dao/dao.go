//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DB

package dao

import (
	"context"
	"errors"
	daoErrors "github.com/gururuby/shortener/internal/domain/dao/errors"
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
)

type DB interface {
	Find(context.Context, string) (*entity.ShortURL, error)
	Save(context.Context, *entity.ShortURL) (*entity.ShortURL, error)
	Ping(context.Context) error
}

type Generator interface {
	UUID() string
	Alias() string
}

type DAO struct {
	gen Generator
	db  DB
}

func New(gen Generator, db DB) *DAO {
	dao := &DAO{
		gen: gen,
		db:  db,
	}

	return dao
}

func (dao *DAO) FindByAlias(ctx context.Context, alias string) (*entity.ShortURL, error) {
	return dao.db.Find(ctx, alias)
}

func (dao *DAO) Save(ctx context.Context, sourceURL string) (*entity.ShortURL, error) {
	shortURL := entity.NewShortURL(dao.gen, sourceURL)
	res, err := dao.db.Save(ctx, shortURL)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBIsNotUnique) {
			return res, daoErrors.ErrDAORecordIsNotUnique
		}
	}
	return res, err
}

func (dao *DAO) IsDBReady(ctx context.Context) error {
	return dao.db.Ping(ctx)
}
