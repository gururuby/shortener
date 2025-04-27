package usecase

import "errors"

var (
	ErrAppDBIsNotReady = errors.New("db is not ready")

	ErrShortURLEmptyBaseURL      = errors.New("empty base URL, please specify base URL")
	ErrShortURLEmptySourceURL    = errors.New("empty source URL, please specify source URL")
	ErrShortURLEmptyAlias        = errors.New("empty alias, please specify alias")
	ErrShortURLSourceURLNotFound = errors.New("source URL not found")
)
