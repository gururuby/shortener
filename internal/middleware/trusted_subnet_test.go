package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		trustedSubnet  string
		clientIP       string
		expectedStatus int
	}{
		{
			name:           "empty subnet denies all",
			path:           "/api/internal/stats",
			method:         http.MethodGet,
			trustedSubnet:  "",
			clientIP:       "192.168.1.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "valid ip in subnet",
			path:           "/api/internal/stats",
			method:         http.MethodGet,
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "192.168.1.10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid ip outside subnet",
			path:           "/api/internal/stats",
			method:         http.MethodGet,
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid ip outside subnet but endpoint is not guarded",
			path:           "/api/internal/stats_unguarded",
			method:         http.MethodGet,
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "10.0.0.1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := TrustedSubnetMiddleware(tt.trustedSubnet)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Real-IP", tt.clientIP)

			rr := httptest.NewRecorder()
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
