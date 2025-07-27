package validator

import (
	"testing"
)

func TestIsInvalidURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		comment string
		want    bool
	}{
		{
			name:    "valid http URL",
			url:     "http://example.com",
			want:    false,
			comment: "Standard HTTP URL should be valid",
		},
		{
			name:    "valid https URL",
			url:     "https://example.com",
			want:    false,
			comment: "Standard HTTPS URL should be valid",
		},
		{
			name:    "valid URL with www",
			url:     "https://www.example.com",
			want:    false,
			comment: "URL with www subdomain should be valid",
		},
		{
			name:    "valid URL with port",
			url:     "http://localhost:8080",
			want:    false,
			comment: "URL with port number should be valid",
		},
		{
			name:    "valid URL with path",
			url:     "https://example.com/path/to/resource",
			want:    false,
			comment: "URL with path should be valid",
		},
		{
			name:    "valid URL with query params",
			url:     "http://example.com?param=value",
			want:    false,
			comment: "URL with query parameters should be valid",
		},
		{
			name:    "invalid missing protocol",
			url:     "example.com",
			want:    true,
			comment: "URL without protocol should be invalid",
		},
		{
			name:    "invalid wrong protocol",
			url:     "ftp://example.com",
			want:    true,
			comment: "URL with non-HTTP/HTTPS protocol should be invalid",
		},
		{
			name:    "empty string",
			url:     "",
			want:    true,
			comment: "Empty string should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvalidURL(tt.url); got != tt.want {
				t.Errorf("IsInvalidURL(%q) = %v, want %v (%s)", tt.url, got, tt.want, tt.comment)
			}
		})
	}
}
