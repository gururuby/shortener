package memory

import (
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
)

type DB struct {
	shortURLs map[string]*entity.ShortURL
}

func New() *DB {
	return &DB{
		shortURLs: make(map[string]*entity.ShortURL),
	}
}

func (db *DB) Find(alias string) (*entity.ShortURL, error) {
	shortURL, ok := db.shortURLs[alias]
	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *DB) findBySourceURL(sourceURL string) (*entity.ShortURL, error) {
	var (
		shortURL  *entity.ShortURL
		noRecords = true
	)

	for _, url := range db.shortURLs {
		if url.SourceURL == sourceURL {
			shortURL = url
			noRecords = false
			break
		}
	}

	if noRecords {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *DB) Save(shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	existRecord, _ := db.findBySourceURL(shortURL.SourceURL)
	if existRecord != nil {
		return existRecord, dbErrors.ErrDBIsNotUnique
	}

	db.shortURLs[shortURL.Alias] = shortURL
	return shortURL, nil
}

func (db *DB) Ping() error {
	return nil
}

func (db *DB) Truncate() {
	db.shortURLs = make(map[string]*entity.ShortURL)
}
