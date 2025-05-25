//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DB

package storage

import (
	"context"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
)

type UserDB interface {
	FindUser(ctx context.Context, id int) (*userEntity.User, error)
	FindUserURLs(ctx context.Context, id int) ([]*shortURLEntity.ShortURL, error)
	SaveUser(ctx context.Context) (*userEntity.User, error)
	MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error
}

type UserStorage struct {
	db UserDB
}

func Setup(db UserDB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) FindURLs(ctx context.Context, id int) ([]*shortURLEntity.ShortURL, error) {
	return s.db.FindUserURLs(ctx, id)
}

func (s *UserStorage) MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error {
	return s.db.MarkURLAsDeleted(ctx, userID, aliases)
}

func (s *UserStorage) FindUser(ctx context.Context, id int) (*userEntity.User, error) {
	return s.db.FindUser(ctx, id)
}

func (s *UserStorage) SaveUser(ctx context.Context) (*userEntity.User, error) {
	return s.db.SaveUser(ctx)
}
