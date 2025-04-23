package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
					AliasLength:           5,
					Env:                   "development",
					MaxGenerationAttempts: 5,
					Name:                  "Shortener",
					Version:               "0.0.1",
					BaseURL:               "http://localhost:8080",
				},
				Server: Server{
					Address: "localhost:8080",
				},
				DB: DB{
					Type: "file",
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
