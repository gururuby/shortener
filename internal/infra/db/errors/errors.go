package errors

import "errors"

var (
	ErrDBIsNotHealthy    = errors.New("db is not healthy")
	ErrDBRecordNotFound  = errors.New("record not found")
	ErrDBIsNotUnique     = errors.New("record is not unique")
	ErrDBRestoreFromFile = errors.New("cannot restore records from file %s")
)
