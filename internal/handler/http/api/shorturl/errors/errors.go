// Package handler contains HTTP request handlers and API endpoint logic.
// It defines API-specific error conditions that clients may encounter.
package handler

import "errors"

// Errors list
var (
	// ErrAPIEmptyBatch indicates a batch API request was received with no items to process.
	//
	// Common scenarios:
	// - Empty array in JSON payload
	//
	// Client handling recommendations:
	// - Verify request payload contains data
	//
	// Example:
	//  POST /api/shorten/batch
	//  Body: []  // Triggers this error
	ErrAPIEmptyBatch = errors.New("nothing to process, empty batch")
)
