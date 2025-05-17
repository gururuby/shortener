package jwt

import "errors"

var (
	ErrJWTUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrJWTTokenInvalid            = errors.New("invalid token")
	ErrJWTParseError              = errors.New("cannot parse jwt")
	ErrJWTCannotSignData          = errors.New("cannot sign data")
)
