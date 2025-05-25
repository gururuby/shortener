package db

import (
	"context"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
)

type MemoryDB struct {
	shortURLs map[string]*shortURLEntity.ShortURL
	users     map[int]*userEntity.User
}

func New() *MemoryDB {
	return &MemoryDB{
		shortURLs: make(map[string]*shortURLEntity.ShortURL),
		users:     make(map[int]*userEntity.User),
	}
}

func (db *MemoryDB) FindUser(_ context.Context, id int) (*userEntity.User, error) {
	user, ok := db.users[id]
	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}
	return user, nil
}

func (db *MemoryDB) FindUserURLs(_ context.Context, userID int) ([]*shortURLEntity.ShortURL, error) {
	var urls []*shortURLEntity.ShortURL

	for _, url := range db.shortURLs {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

func (db *MemoryDB) SaveUser(_ context.Context) (*userEntity.User, error) {
	id := len(db.users) + 1
	user := &userEntity.User{ID: id}
	db.users[id] = user
	return user, nil
}

func (db *MemoryDB) FindShortURL(_ context.Context, alias string) (*shortURLEntity.ShortURL, error) {
	shortURL, ok := db.shortURLs[alias]
	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *MemoryDB) MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error {
	return nil
}

func (db *MemoryDB) findShortURLBySourceURL(_ context.Context, sourceURL string) (*shortURLEntity.ShortURL, error) {
	var (
		shortURL  *shortURLEntity.ShortURL
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

func (db *MemoryDB) SaveShortURL(ctx context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error) {
	existRecord, _ := db.findShortURLBySourceURL(ctx, shortURL.SourceURL)
	if existRecord != nil {
		return existRecord, dbErrors.ErrDBIsNotUnique
	}

	db.shortURLs[shortURL.Alias] = shortURL
	return shortURL, nil
}

func (db *MemoryDB) Ping(_ context.Context) error {
	return nil
}
