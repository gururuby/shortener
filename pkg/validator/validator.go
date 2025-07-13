/*
Package validator provides input validation utilities for the application.

It includes functions for validating common data formats like URLs.
*/
package validator

import "regexp"

// IsInvalidURL checks if a string is not a valid HTTP/HTTPS URL.
// It validates the URL format using a regular expression that matches:
//   - http:// or https:// protocols
//   - Optional www. subdomain
//   - Domain names with word characters
//   - Optional port numbers
//   - Optional path/query parameters
//
// Parameters:
//   - rawURL: The URL string to validate
//
// Returns:
//   - bool: true if the URL is invalid, false if valid
//
// Example:
//
//	if validator.IsInvalidURL("https://example.com") {
//	    // handle invalid URL
//	}
func IsInvalidURL(rawURL string) bool {
	reg := regexp.MustCompile(`\Ahttps?://(www\.)?\w+(:\d{1,5})?\.?(\w+)?.*\z`)
	return !reg.MatchString(rawURL)
}
