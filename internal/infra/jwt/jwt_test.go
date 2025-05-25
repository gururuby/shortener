package jwt

import (
	jwtErrors "github.com/gururuby/shortener/internal/infra/jwt/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
	"time"
)

func TestJWT_SignUserID(t *testing.T) {
	var tests = []struct {
		name    string
		secret  string
		expTime time.Duration
	}{
		{
			name:    "success generation",
			secret:  "secret",
			expTime: 10 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt := New(tt.secret, tt.expTime)
			token, err := jwt.SignUserID(1)
			require.NoError(t, err)
			assert.Regexp(t, regexp.MustCompile(`.+\..+\..+`), token)
		})
	}
}

func TestJWT_ReadUserID_OK(t *testing.T) {
	var tests = []struct {
		name    string
		secret  string
		expTime time.Duration
		userID  int
	}{
		{
			name:    "success reading user ID",
			secret:  "secret",
			expTime: 10 * time.Minute,
			userID:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				err   error
				token string
				id    int
			)
			jwt := New(tt.secret, tt.expTime)
			token, err = jwt.SignUserID(tt.userID)
			require.NoError(t, err)
			id, err = jwt.ReadUserID(token)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, id)
		})
	}
}

func TestJWT_ReadUserID_Errors(t *testing.T) {
	var tests = []struct {
		name    string
		token   string
		secret  string
		expTime time.Duration
		userID  int
		err     error
	}{
		{
			name:    "when passed incorrect token",
			token:   "incorrect token",
			secret:  "secret",
			expTime: 10 * time.Minute,
			userID:  1,
			err:     jwtErrors.ErrJWTParseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt := New(tt.secret, tt.expTime)
			_, err := jwt.ReadUserID(tt.token)
			require.ErrorIs(t, err, tt.err)
		})
	}
}
