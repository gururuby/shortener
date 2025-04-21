//go:generate mockgen -destination=./mock_usecase/mock.go . DAO

package usecase

import (
	"github.com/gururuby/shortener/internal/domain/entity"
	ucErrors "github.com/gururuby/shortener/internal/domain/usecase/errors"
	"strings"
)

type DAO interface {
	FindByAlias(alias string) (*entity.ShortURL, error)
	Save(sourceURL string) (*entity.ShortURL, error)
}

type UseCase struct {
	baseURL string
	dao     DAO
}

func NewUseCase(dao DAO, baseURL string) *UseCase {
	return &UseCase{
		dao:     dao,
		baseURL: baseURL,
	}
}

func (u *UseCase) CreateShortURL(sourceURL string) (string, error) {
	if u.baseURL == "" {
		return "", ucErrors.ErrEmptyBaseURL
	}

	if sourceURL == "" {
		return "", ucErrors.ErrEmptySourceURL
	}

	result, err := u.dao.Save(sourceURL)

	if err != nil {
		return "", err
	}

	return u.baseURL + "/" + result.Alias, nil
}

func (u *UseCase) FindShortURL(alias string) (string, error) {
	alias = strings.TrimPrefix(alias, "/")

	if alias == "" {
		return "", ucErrors.ErrEmptyAlias
	}

	res, err := u.dao.FindByAlias(alias)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", ucErrors.ErrSourceURLNotFound
	}

	return res.SourceURL, nil
}
