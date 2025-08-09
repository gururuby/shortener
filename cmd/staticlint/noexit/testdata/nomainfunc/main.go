// Package main demonstrates a Go package with os.Exit usage that would be flagged
// by static analysis tools enforcing proper exit handling.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("os.Exit calls in another function")
}

// Foo func
func Foo() {
	os.Exit(1)
}
