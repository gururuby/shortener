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
					AliasLength:     5,
					Env:             "development",
					Name:            "Shortener",
					ShutdownTimeout: 30 * time.Second,
					Version:         "0.0.1",
					BaseURL:         "http://localhost:8080",
				},
				Auth: Auth{
					TokenTTL:  24 * time.Hour,
					SecretKey: "secret",
				},
				Server: Server{
					Address:       "localhost:8080",
					ReadTimeout:   5 * time.Second,
					TrustedSubnet: "127.0.0.1/24",
					WriteTimeout:  10 * time.Second,
					IdleTimeout:   120 * time.Second,
					HTTPS: HTTPS{
						Enabled: false,
					},
					GRPC: GRPC{
						Enabled:               false,
						ConnectionTimeout:     120 * time.Second,
						Address:               ":50051",
						MaxConnectionIdle:     2 * time.Hour,
						MaxConnectionAge:      30 * time.Minute,
						MaxConnectionAgeGrace: 5 * time.Minute,
						KeepaliveTime:         2 * time.Hour,
						KeepaliveTimeout:      20 * time.Second,
						MinKeepaliveTime:      10 * time.Second,
						PermitWithoutStream:   true,
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
