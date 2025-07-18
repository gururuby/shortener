// Package jwt provides JSON Web Token (JWT) creation and validation utilities.
// It defines common JWT processing errors for secure authentication flows.
package jwt

import "errors"

// Errors list
var (
	// ErrJWTUnexpectedSigningMethod indicates the token uses an unexpected
	// signing algorithm that doesn't match the expected method.
	//
	// Security implications:
	// - Potential token forgery attempt
	// - Algorithm downgrade attack
	//
	// Recommended action:
	// - Reject token immediately
	// - Log the incident for security monitoring
	// - Verify token issuer configuration
	ErrJWTUnexpectedSigningMethod = errors.New("unexpected signing method")

	// ErrJWTTokenInvalid indicates the token failed validation checks.
	//
	// Common causes:
	// - Malformed token structure
	// - Invalid signature
	// - Missing required claims
	//
	// Handling guidance:
	// - Return HTTP 401 Unauthorized
	// - Clear client-side authentication state
	// - Do not expose validation details to clients
	ErrJWTTokenInvalid = errors.New("invalid token")

	// ErrJWTParseError indicates failure to parse the JWT string.
	//
	// Typical scenarios:
	// - Incorrect token format
	// - Base64 decoding failure
	// - JSON unmarshalling error
	//
	// Debugging tips:
	// - Verify token is complete (3 parts separated by dots)
	// - Check for URL-safe base64 encoding
	// - Validate header JSON structure
	ErrJWTParseError = errors.New("cannot parse jwt")

	// ErrJWTCannotSignData indicates failure during token signing.
	//
	// Possible reasons:
	// - Invalid signing key
	// - Unsupported algorithm
	// - Key/algorithm mismatch
	//
	// Resolution steps:
	// - Verify key format and type
	// - Check algorithm compatibility
	// - Ensure proper key initialization
	ErrJWTCannotSignData = errors.New("cannot sign data")
)
