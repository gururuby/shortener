//go:generate mockgen -destination=./mocks/mock.go -package=mocks . DAO

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
	if u.baseURL == "" {
		return "", ucErrors.ErrShortURLEmptyBaseURL
	}

	if sourceURL == "" {
		return "", ucErrors.ErrShortURLEmptySourceURL
	}

	result, err := u.dao.Save(sourceURL)

	if err != nil {
		return "", err
	}

	return u.baseURL + "/" + result.Alias, nil
}

func (u ShortURLUseCase) FindShortURL(alias string) (string, error) {
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
