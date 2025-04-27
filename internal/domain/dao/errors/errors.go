package usecase

import "errors"

var (
	ErrDAORecordNotFound    = errors.New("record not found")
	ErrDAORecordIsNotUnique = errors.New("record is not unique")
	ErrDAORestoreFromFile   = errors.New("cannot restore records from file %s")
	ErrDAOIsNotReadyDB      = errors.New("database is not ready")
)
