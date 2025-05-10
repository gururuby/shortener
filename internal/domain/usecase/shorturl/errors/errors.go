package usecase

import "errors"

var (
	ErrShortURLAlreadyExist      = errors.New("short URL already exist")
	ErrShortURLInvalidBaseURL    = errors.New("invalid base URL, please specify valid URL")
	ErrShortURLInvalidSourceURL  = errors.New("invalid source URL, please specify valid URL")
	ErrShortURLEmptyAlias        = errors.New("empty alias, please specify alias")
	ErrShortURLSourceURLNotFound = errors.New("source URL not found")
)
