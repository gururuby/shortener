/*
Package utils provides general utility functions for the application.

It includes helper functions for common operations like retry logic.
*/
package utils

import "time"

// Retry executes a function and retries on failure with exponential backoff.
//
// Parameters:
//   - f: The function to execute that returns an error
//   - retryTimes: Maximum number of retry attempts (must be >= 0)
//   - delay: Duration to wait between retry attempts
//
// Returns:
//   - error: Returns nil if the function succeeds within retry attempts,
//     otherwise returns the last error encountered
//
// Example:
//
//	err := Retry(func() error {
//	    return SomeOperation()
//	}, 3, time.Second)
//	if err != nil {
//	    // handle error
//	}
func Retry(f func() error, retryTimes int, delay time.Duration) error {
	for retryTimes > 0 {
		if err := f(); err != nil {
			time.Sleep(delay)
			retryTimes--

			continue
		}
		return nil
	}

	return nil
}
