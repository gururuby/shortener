// Package errors defines database-related error conditions that may occur
// during application data operations. These errors provide consistent
// error handling across different database implementations.
package errors

import "errors"

// Errors list
var (
	// ErrDBIsNotHealthy indicates the database connection is unavailable
	// or not responding to health checks.
	//
	// Typical causes:
	// - Database server is down
	// - Network connectivity issues
	// - Connection pool exhausted
	//
	// Recommended actions:
	// - Check database server status
	// - Verify connection parameters
	// - Implement retry logic with backoff
	ErrDBIsNotHealthy = errors.New("db is not healthy")

	// ErrDBRecordNotFound indicates a requested record does not exist
	// in the database.
	//
	// Common scenarios:
	// - Invalid record ID lookup
	// - Record was deleted
	// - Table is empty
	//
	// Handling suggestions:
	// - Return HTTP 404 for API responses
	// - Consider upsert operations where appropriate
	ErrDBRecordNotFound = errors.New("record not found")

	// ErrDBQuery indicates a database query failed to execute.
	//
	// Possible reasons:
	// - Syntax error in query
	// - Permission denied
	// - Connection interrupted
	//
	// Debugging tips:
	// - Check query logs
	// - Verify table schema matches query
	// - Test query in database client
	ErrDBQuery = errors.New("query to db is failed")

	// ErrDBIsNotUnique indicates a uniqueness constraint violation.
	//
	// Common cases:
	// - Duplicate primary key
	// - Unique index violation
	// - Concurrent insert of same data
	//
	// Resolution options:
	// - Use different unique value
	// - Implement upsert logic
	// - Check for race conditions
	ErrDBIsNotUnique = errors.New("record is not unique")

	// ErrDBRestoreFromFile indicates failure during database restoration
	// from a backup file.
	//
	// Format note:
	// - Use fmt.Errorf to include filename: fmt.Errorf(ErrDBRestoreFromFile.Error(), "backup.json")
	//
	// Potential issues:
	// - Invalid file format
	// - File permissions
	// - Schema version mismatch
	ErrDBRestoreFromFile = errors.New("cannot restore records from file %s")
)
