package errors

import "errors"

var (
	ErrGeneratorEmptyAliasLength = errors.New("alias length is zero, please configure correct value")
)
