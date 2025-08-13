/*
Package jwt implements JSON Web Token (JWT) creation and validation.

It provides:
- JWT generation with user claims
- Token signing and verification
- Configurable token expiration
- Custom error handling for JWT operations
*/
package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	jwtErrors "github.com/gururuby/shortener/internal/infra/jwt/errors"
)

// claims contains the JWT claims structure including registered claims
// and custom user ID field.
type claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"` // User ID to be stored in the token
}

// JWT provides methods for creating and validating JWT tokens.
type JWT struct {
	secret   []byte        // Secret key used for signing tokens
	tokenTTL time.Duration // Token time-to-live duration
}

// New creates a new JWT instance with the given secret and token TTL.
// Parameters:
// - secret: Secret key for signing tokens
// - ttl: Duration until token expiration
// Returns:
// - *JWT: Initialized JWT instance
func New(secret string, ttl time.Duration) *JWT {
	return &JWT{secret: []byte(secret), tokenTTL: ttl}
}

// SignUserID creates a new JWT token containing the user ID.
// Parameters:
// - userID: User ID to embed in the token
// Returns:
// - string: Signed JWT token
// - error: jwtErrors.ErrJWTCannotSignData if signing fails
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

// ReadUserID validates a JWT token and extracts the user ID.
// Parameters:
// - tokenString: JWT token to validate
// Returns:
// - int: User ID extracted from the token
// - error: Various JWT validation errors if token is invalid
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
