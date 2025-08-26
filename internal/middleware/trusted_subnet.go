// Package middleware provides HTTP middleware components for request processing.
// It includes security and validation middleware for the application.
package middleware

import (
	"net"
	"net/http"
)

var (
	// guardEndpoints contains the list of endpoints that require IP validation.
	guardEndpoints = []endpoint{
		{
			path:   "/api/internal/stats",
			method: http.MethodGet,
		},
	}
)

// endpoint represents an HTTP endpoint with path and method.
type endpoint struct {
	path   string // HTTP path
	method string // HTTP method
}

// TrustedSubnetMiddleware creates middleware that validates client IP against a trusted subnet.
// It only protects endpoints defined in guardEndpoints. Returns 403 Forbidden if validation fails.
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isGuardEndpoint(r) {
				next.ServeHTTP(w, r)
				return
			}

			if trustedSubnet == "" {
				http.Error(w, "Access forbidden", http.StatusForbidden)
				return
			}

			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				http.Error(w, "X-Real-IP header required", http.StatusForbidden)
				return
			}

			if !isIPInSubnet(clientIP, trustedSubnet) {
				http.Error(w, "Access forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isGuardEndpoint checks if the request targets a protected endpoint.
func isGuardEndpoint(r *http.Request) bool {
	for _, e := range guardEndpoints {
		if e.path == r.URL.Path && e.method == r.Method {
			return true
		}
	}
	return false
}

// isIPInSubnet validates if an IP address belongs to the specified subnet.
// Returns false if IP or subnet parsing fails.
func isIPInSubnet(ipStr, subnetStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, subnet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return false
	}

	return subnet.Contains(ip)
}
