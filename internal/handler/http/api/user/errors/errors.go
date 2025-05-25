package handler

import "errors"

var (
	ErrHandlerNoAliasesForDelete = errors.New("no aliases passed to delete short urls")
)
