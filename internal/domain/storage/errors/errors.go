// Package errors defines common storage-related error variables for consistent error handling.
// These errors are typically returned by storage implementations and handled by higher layers.
package errors

import "errors"

// Errors list
var (
	// ErrStorageRecordNotFound indicates that a requested record could not be found in storage.
	// This error should be returned when querying for a non-existent resource.
	ErrStorageRecordNotFound = errors.New("record not found")

	// ErrStorageRecordIsNotUnique indicates a violation of uniqueness constraints.
	// This error should be returned when attempting to create or update a record with duplicate values
	// in a field that requires uniqueness (e.g., duplicate short URL alias).
	ErrStorageRecordIsNotUnique = errors.New("record is not unique")

	// ErrStorageIsNotReadyDB indicates that the database connection or storage system isn't ready.
	// This error should be returned during health checks or when the storage backend is unavailable.
	ErrStorageIsNotReadyDB = errors.New("database is not ready")
)
