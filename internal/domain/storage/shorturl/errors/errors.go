package storage

import "errors"

var (
	ErrStorageRecordNotFound    = errors.New("record not found")
	ErrStorageRecordIsNotUnique = errors.New("record is not unique")
	ErrStorageIsNotReadyDB      = errors.New("database is not ready")
)
