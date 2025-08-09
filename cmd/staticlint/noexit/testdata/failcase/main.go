// Package main demonstrates a violation of the noexit analyzer rule.
//
// This file serves as an example of what NOT to do - it contains a direct call
// to os.Exit in the main function, which would be flagged by the noexit analyzer.
//
// The analyzer would report:
//
//	main.go:6:2: direct call to os.Exit in main function of main package is forbidden
//
// This example exists purely for testing purposes to verify that the noexit analyzer
// correctly detects and reports this pattern.
//
// Correct usage would involve:
//   - Returning an error code from main instead
//   - Using proper error handling
//   - Allowing deferred functions to run before exit
//
// Note: This file would typically be placed in testdata directory for analyzer testing,
// not in actual production code.
package main

import "os"

func main() {
	os.Exit(1) // want "direct call to os.Exit in main function of main package is forbidden"
}
