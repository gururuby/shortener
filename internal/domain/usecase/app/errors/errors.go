package usecase

import "errors"

var (
	ErrAppDBIsNotReady = errors.New("db is not ready")
)
