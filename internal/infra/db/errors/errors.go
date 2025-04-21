package errors

import "errors"

var (
	ErrNotFound        = errors.New("record not found")
	ErrIsNotUnique     = errors.New("record is not unique")
	ErrRestoreFromFile = errors.New("cannot restore records from file %s")
)
