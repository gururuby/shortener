package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *Config
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
				Server: Server{
					Address: "localhost:8080",
				},
				Database: Database{
					Type:         "postgresql",
					DSN:          "postgresql://postgres:pass@0.0.0.0:5432/shortener?sslmode=disable",
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
