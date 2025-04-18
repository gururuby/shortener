//go:generate mockgen -destination=./mock_usecase/mock.go . DAO

package usecase

import (
	"errors"
	"github.com/gururuby/shortener/internal/domain/entity"
	"strings"
)

const (
	EmptyBaseURLError      = "empty base URL, please specify base URL"
	EmptySourceURLError    = "empty source URL, please specify source URL"
	EmptyAliasError        = "empty alias, please specify alias"
	SourceURLNotFoundError = "source URL not found"
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
		return "", errors.New(EmptyBaseURLError)
	}

	if sourceURL == "" {
		return "", errors.New(EmptySourceURLError)
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
		return "", errors.New(EmptyAliasError)
	}

	res, err := u.dao.FindByAlias(alias)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", errors.New(SourceURLNotFoundError)
	}

	return res.SourceURL, nil
}
