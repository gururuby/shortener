package usecase

import (
	"errors"
	"strings"
)

const (
	EmptyBaseURLError      = "empty base URL, please specify base URL"
	EmptySourceURLError    = "empty source URL, please specify source URL"
	EmptyAliasError        = "empty alias, please specify alias"
	SourceURLNotFoundError = "source URL not found"
)

type shortURLDAO interface {
	FindByAlias(alias string) (string, error)
	Save(sourceURL string) (string, error)
}

type ShortURLUseCase struct {
	baseURL string
	dao     shortURLDAO
}

func NewShortURLUseCase(dao shortURLDAO, baseURL string) *ShortURLUseCase {
	return &ShortURLUseCase{
		dao:     dao,
		baseURL: baseURL,
	}
}

func (u *ShortURLUseCase) CreateShortURL(sourceURL string) (string, error) {
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

	return u.baseURL + "/" + result, nil
}

func (u *ShortURLUseCase) FindShortURL(alias string) (string, error) {
	alias = strings.TrimPrefix(alias, "/")

	if alias == "" {
		return "", errors.New(EmptyAliasError)
	}

	res, err := u.dao.FindByAlias(alias)
	if err != nil {
		return "", err
	}

	if res == "" {
		return "", errors.New(SourceURLNotFoundError)
	}

	return res, nil
}
