//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DB

package dao

import (
	"errors"
	daoErrors "github.com/gururuby/shortener/internal/domain/dao/errors"
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
)

type DB interface {
	Find(string) (*entity.ShortURL, error)
	Save(*entity.ShortURL) (*entity.ShortURL, error)
	Ping() error
	Truncate()
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

func (dao *DAO) FindByAlias(alias string) (*entity.ShortURL, error) {
	return dao.db.Find(alias)
}

func (dao *DAO) Save(sourceURL string) (*entity.ShortURL, error) {
	shortURL := entity.NewShortURL(dao.gen, sourceURL)
	res, err := dao.db.Save(shortURL)
	if err != nil {
		if errors.Is(err, dbErrors.ErrDBIsNotUnique) {
			return res, daoErrors.ErrDAORecordIsNotUnique
		}
	}
	return res, err
}

func (dao *DAO) IsDBReady() error {
	return dao.db.Ping()
}

func (dao *DAO) Clear() {
	dao.db.Truncate()
}
