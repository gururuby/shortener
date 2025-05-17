package validator

import "testing"

func TestIsInvalidURL(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		want   bool
	}{
		{
			name:   "when passed valid URL",
			rawURL: "http://www.example.com",
			want:   false,
		},
		{
			name:   "when passed valid URL without www",
			rawURL: "http://example.com",
			want:   false,
		},
		{
			name:   "when passed valid URL with port",
			rawURL: "http://example.com:80",
			want:   false,
		},
		{
			name:   "when passed valid localhost URL with port",
			rawURL: "http://localhost:80",
			want:   false,
		},
		{
			name:   "when passed invalid URL with incorrect protocol",
			rawURL: "ttp://example.com",
			want:   true,
		},
		{
			name:   "when passed invalid URL without protocol",
			rawURL: "example.com",
			want:   true,
		},
		{
			name:   "when passed empty URL",
			rawURL: "example.com",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvalidURL(tt.rawURL); got != tt.want {
				t.Errorf("IsInvalidURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
