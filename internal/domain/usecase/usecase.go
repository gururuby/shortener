//go:generate mockgen -destination=./mock_usecase/mock.go . DAO

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

type DAO interface {
	FindByAlias(alias string) (string, error)
	Save(sourceURL string) (string, error)
}

type UseCase struct {
	baseURL string
	DAO     DAO
}

func NewUseCase(dao DAO, baseURL string) *UseCase {
	return &UseCase{
		DAO:     dao,
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

	result, err := u.DAO.Save(sourceURL)

	if err != nil {
		return "", err
	}

	return u.baseURL + "/" + result, nil
}

func (u *UseCase) FindShortURL(alias string) (string, error) {
	alias = strings.TrimPrefix(alias, "/")

	if alias == "" {
		return "", errors.New(EmptyAliasError)
	}

	res, err := u.DAO.FindByAlias(alias)
	if err != nil {
		return "", err
	}

	if res == "" {
		return "", errors.New(SourceURLNotFoundError)
	}

	return res, nil
}
