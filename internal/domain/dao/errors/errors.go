package usecase

import "errors"

var (
	ErrDAORecordNotFound    = errors.New("record not found")
	ErrDAORecordIsNotUnique = errors.New("record is not unique")
	ErrDAOIsNotReadyDB      = errors.New("database is not ready")
)
