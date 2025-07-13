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

// main is the application entry point.
//
// It performs:
//  1. Configuration initialization
//  2. Application instance creation and setup
//  3. HTTP server startup
//
// If any step fails, it logs the error and terminates.
func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("cannot setup config: %s", err)
	}
	app.New(cfg).Setup().Run()
}
