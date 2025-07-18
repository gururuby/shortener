// Package usecase contains core business logic for user authentication and management.
// It defines domain-specific errors for user-related operations.
package usecase

import "errors"

// Errors list
var (
	// ErrUserCannotAuthenticate indicates failure during user authentication.
	//
	// Common causes:
	// - Invalid credentials (email/password)
	// - Account locked/suspended
	// - Expired session/token
	//
	// Security considerations:
	// - Should not reveal specific failure reason
	// - Should implement rate limiting
	ErrUserCannotAuthenticate = errors.New("cannot authenticate user")

	// ErrUserNotFound indicates no user exists with the provided identifier.
	//
	// Typical scenarios:
	// - Invalid user ID lookup
	// - Email not registered
	// - Deleted account
	//
	// Privacy note:
	// - Can be used to probe for registered emails
	// - Consider generic messaging in UI
	ErrUserNotFound = errors.New("user is not found")

	// ErrUserCannotSave indicates failure persisting user data changes.
	//
	// Common root causes:
	// - Database constraints violation
	// - Invalid field values
	// - Storage system failure
	//
	// Recovery suggestions:
	// - Validate data before saving
	// - Implement retry logic
	ErrUserCannotSave = errors.New("cannot save user")

	// ErrUserCannotRegister indicates failure during new user registration.
	//
	// Specific cases:
	// - Duplicate email/username
	// - Invalid profile data
	// - Failed dependency (e.g., email service)
	//
	// UX recommendations:
	// - Provide specific validation feedback
	// - Suggest alternative usernames if conflict
	ErrUserCannotRegister = errors.New("cannot register user")

	// ErrUserStorageNotWorking indicates critical failure in user data storage.
	//
	// System implications:
	// - All user operations will fail
	// - Requires immediate attention
	//
	// Monitoring alerts:
	// - Should trigger high-priority alerts
	// - May require manual intervention
	ErrUserStorageNotWorking = errors.New("user storage is not working")
)
