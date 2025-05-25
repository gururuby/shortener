package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	jwtErrors "github.com/gururuby/shortener/internal/infra/jwt/errors"
	"time"
)

type claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

type JWT struct {
	secret   []byte
	tokenTTL time.Duration
}

func New(secret string, ttl time.Duration) *JWT {
	return &JWT{secret: []byte(secret), tokenTTL: ttl}
}

func (j *JWT) SignUserID(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenTTL)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", jwtErrors.ErrJWTCannotSignData
	}

	return tokenString, nil
}

func (j *JWT) ReadUserID(tokenString string) (int, error) {
	clms := &claims{}
	token, err := jwt.ParseWithClaims(tokenString, clms,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwtErrors.ErrJWTUnexpectedSigningMethod
			}
			return j.secret, nil
		})
	if err != nil {
		return 0, jwtErrors.ErrJWTParseError
	}

	if !token.Valid {
		return 0, jwtErrors.ErrJWTTokenInvalid
	}

	return clms.UserID, nil

}
