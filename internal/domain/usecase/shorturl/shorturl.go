//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DAO

package usecase

import (
	"errors"
	daoErrors "github.com/gururuby/shortener/internal/domain/dao/errors"
	"github.com/gururuby/shortener/internal/domain/entity"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"github.com/gururuby/shortener/internal/infra/logger"
	"github.com/gururuby/shortener/internal/infra/utils/validator"
	"strings"
)

type DAO interface {
	FindByAlias(alias string) (*entity.ShortURL, error)
	Save(sourceURL string) (*entity.ShortURL, error)
}

type ShortURLUseCase struct {
	baseURL string
	dao     DAO
}

func NewShortURLUseCase(dao DAO, baseURL string) *ShortURLUseCase {
	return &ShortURLUseCase{
		dao:     dao,
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

	result, err := u.dao.Save(sourceURL)

	if err != nil {
		if errors.Is(err, daoErrors.ErrDAORecordIsNotUnique) {
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

	res, err := u.dao.FindByAlias(alias)
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
