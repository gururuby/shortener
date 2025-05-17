package usecase

import "errors"

var (
	ErrUserCannotAuthenticate = errors.New("cannot authenticate user")
	ErrUserNotFound           = errors.New("user is not found")
	ErrUserCannotSave         = errors.New("cannot save user")
	ErrUserCannotRegister     = errors.New("cannot register user")
	ErrUserStorageNotWorking  = errors.New("user storage is not working")
)
