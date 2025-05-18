//go:generate mockgen -destination=./mocks/mock.go -package=mocks . ShortURLStorage

package usecase

import (
	"context"
	"errors"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/errors"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/pkg/validator"
	"strings"
)

type ShortURLStorage interface {
	FindShortURL(ctx context.Context, alias string) (*entity.ShortURL, error)
	SaveShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (*entity.ShortURL, error)
}

type ShortURLUseCase struct {
	baseURL string
	storage ShortURLStorage
}

func NewShortURLUseCase(storage ShortURLStorage, baseURL string) *ShortURLUseCase {
	return &ShortURLUseCase{
		storage: storage,
		baseURL: baseURL,
	}
}

func (u *ShortURLUseCase) CreateShortURL(ctx context.Context, user *userEntity.User, sourceURL string) (string, error) {
	if validator.IsInvalidURL(u.baseURL) {
		return "", ucErrors.ErrShortURLInvalidBaseURL
	}

	if validator.IsInvalidURL(sourceURL) {
		return "", ucErrors.ErrShortURLInvalidSourceURL
	}

	result, err := u.storage.SaveShortURL(ctx, user, sourceURL)

	if err != nil {
		if errors.Is(err, storageErrors.ErrStorageRecordIsNotUnique) {
			return u.baseURL + "/" + result.Alias, ucErrors.ErrShortURLAlreadyExist
		} else {
			return "", err
		}
	}

	return u.baseURL + "/" + result.Alias, nil
}

func (u *ShortURLUseCase) FindShortURL(ctx context.Context, alias string) (string, error) {
	alias = strings.TrimPrefix(alias, "/")

	if alias == "" {
		return "", ucErrors.ErrShortURLEmptyAlias
	}

	res, err := u.storage.FindShortURL(ctx, alias)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", ucErrors.ErrShortURLSourceURLNotFound
	}

	if res.IsDeleted {
		return "", ucErrors.ErrShortURLDeleted
	}

	return res.SourceURL, nil
}

func (u *ShortURLUseCase) BatchShortURLs(ctx context.Context, urls []entity.BatchShortURLInput) []entity.BatchShortURLOutput {
	var res []entity.BatchShortURLOutput

	for _, url := range urls {
		shortURL, err := u.CreateShortURL(ctx, nil, url.OriginalURL)
		if err != nil {
			logger.Log.Info(err.Error())
			continue
		}
		res = append(res, entity.BatchShortURLOutput{CorrelationID: url.CorrelationID, ShortURL: shortURL})
	}

	return res
}
