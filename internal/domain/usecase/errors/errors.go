package errors

import "errors"

var (
	ErrEmptyBaseURL      = errors.New("empty base URL, please specify base URL")
	ErrEmptySourceURL    = errors.New("empty source URL, please specify source URL")
	ErrEmptyAlias        = errors.New("empty alias, please specify alias")
	ErrSourceURLNotFound = errors.New("source URL not found")
)
