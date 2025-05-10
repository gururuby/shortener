//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Storage

package usecase

import (
	"errors"
	"github.com/gururuby/shortener/internal/domain/entity/shorturl"
	storageErrors "github.com/gururuby/shortener/internal/domain/storage/shorturl/errors"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/shorturl/errors"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/utils/validator"
	"strings"
)

type Storage interface {
	FindByAlias(alias string) (*entity.ShortURL, error)
	Save(sourceURL string) (*entity.ShortURL, error)
}

type ShortURLUseCase struct {
	baseURL string
	storage Storage
}

func NewShortURLUseCase(storage Storage, baseURL string) *ShortURLUseCase {
	return &ShortURLUseCase{
		storage: storage,
		baseURL: baseURL,
	}
}

func (u *ShortURLUseCase) CreateShortURL(sourceURL string) (string, error) {
	if validator.IsInvalidURL(u.baseURL) {
		return "", ucErrors.ErrShortURLInvalidBaseURL
	}

	if validator.IsInvalidURL(sourceURL) {
		return "", ucErrors.ErrShortURLInvalidSourceURL
	}

	result, err := u.storage.Save(sourceURL)

	if err != nil {
		if errors.Is(err, storageErrors.ErrStorageRecordIsNotUnique) {
			return u.baseURL + "/" + result.Alias, ucErrors.ErrShortURLAlreadyExist
		} else {
			return "", err
		}
	}

	return u.baseURL + "/" + result.Alias, nil
}

func (u *ShortURLUseCase) FindShortURL(alias string) (string, error) {
	alias = strings.TrimPrefix(alias, "/")

	if alias == "" {
		return "", ucErrors.ErrShortURLEmptyAlias
	}

	res, err := u.storage.FindByAlias(alias)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", ucErrors.ErrShortURLSourceURLNotFound
	}

	return res.SourceURL, nil
}

func (u *ShortURLUseCase) BatchShortURLs(urls []entity.BatchShortURLInput) []entity.BatchShortURLOutput {
	var res []entity.BatchShortURLOutput

	for _, url := range urls {
		shortURL, err := u.CreateShortURL(url.OriginalURL)
		if err != nil {
			logger.Log.Info(err.Error())
			continue
		}
		res = append(res, entity.BatchShortURLOutput{CorrelationID: url.CorrelationID, ShortURL: shortURL})
	}

	return res
}
