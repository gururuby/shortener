//go:generate mockgen -destination=./mock_dao/mock.go . DB

package dao

import (
	"errors"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/entity"
)

var (
	errNonUnique = errors.New("record is not unique")
	errSave      = errors.New("cannot save error")
)

type DB interface {
	Find(string) (*entity.ShortURL, error)
	Save(*entity.ShortURL) (*entity.ShortURL, error)
}

type Generator interface {
	UUID() string
	Alias() string
}

type DAO struct {
	gen Generator
	cfg *config.Config
	db  DB
}

func New(gen Generator, cfg *config.Config, db DB) *DAO {
	dao := &DAO{
		gen: gen,
		cfg: cfg,
		db:  db,
	}

	return dao
}

func (dao *DAO) FindByAlias(alias string) (*entity.ShortURL, error) {
	return dao.db.Find(alias)
}

func (dao *DAO) Save(sourceURL string) (*entity.ShortURL, error) {
	return dao.saveWithAttempt(1, sourceURL)
}

func (dao *DAO) saveWithAttempt(startAttemptCount int, sourceURL string) (*entity.ShortURL, error) {
	if startAttemptCount > dao.cfg.App.MaxGenerationAttempts {
		return nil, errSave
	}

	shortURL := entity.NewShortURL(dao.gen, sourceURL)
	record, err := dao.db.Save(shortURL)

	if errors.Is(err, errNonUnique) {
		startAttemptCount++
		return dao.saveWithAttempt(startAttemptCount, sourceURL)
	}

	return record, err
}
