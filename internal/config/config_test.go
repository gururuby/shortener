package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		want    *Config
		name    string
		wantErr bool
	}{
		{
			name: "setup default values",
			want: &Config{
				App: App{
					AliasLength: 5,
					Env:         "development",
					Name:        "Shortener",
					Version:     "0.0.1",
					BaseURL:     "http://localhost:8080",
				},
				Auth: Auth{
					TokenTTL:  24 * time.Hour,
					SecretKey: "secret",
				},
				Server: Server{
					Address: "localhost:8080",
					HTTPS: HTTPS{
						Enabled: false,
					},
				},
				Database: Database{
					Type:         "file",
					DSN:          "",
					ConnTryDelay: 5 * time.Second,
					ConnTryTimes: 5,
				},
				FileStorage: FileStorage{
					Path: "/tmp/db.json",
				},
				Log: Log{
					Level: "info",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New()
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
