/*
Package main is the entry point for the URL shortener service application.

The application provides:
- HTTP API for creating short URLs
- Mapping storage between short and long URLs
- User authentication
- Batch URL processing

Key components:
- Configuration (config)
- Application logic (app)
- HTTP request handlers (handler)
- Data storage (storage)
*/
package main

import (
	"github.com/gururuby/shortener/internal/app"
	"github.com/gururuby/shortener/internal/config"
	"log"
)

// Global variables storing build information.
// These are set during the build process using ldflags.
var (
	buildVersion string // Version number of the build
	buildDate    string // Date when the build was created
	buildCommit  string // Git commit hash of the build
)

// main is the application entry point.
//
// It performs:
//  1. Configuration initialization
//  2. Application instance creation and setup
//  3. HTTP server startup
//
// If any step fails, it logs the error and terminates.
func main() {
	logBuildInfo()
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("cannot setup config: %s", err)
	}
	app.New(cfg).Setup().Run()
}

// logBuildInfo logs the build version, date and commit information.
// If any build information is empty, it will be displayed as "N/A".
func logBuildInfo() {
	log.Printf("Build version: %s", handleBuildValue(buildVersion))
	log.Printf("Build date: %s", handleBuildValue(buildDate))
	log.Printf("Build commit: %s", handleBuildValue(buildCommit))
}

// handleBuildValue returns "N/A" if the input string is empty,
// otherwise returns the string itself.
//
// This is used to handle unset build variables gracefully.
func handleBuildValue(v string) string {
	if v == "" {
		return "N/A"
	}
	return v
}
