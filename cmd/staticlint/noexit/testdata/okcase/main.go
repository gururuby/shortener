// Package main demonstrates a Go package with os.Exit usage that would be flagged
// by static analysis tools enforcing proper exit handling.
package main

import "fmt"

func main() {
	fmt.Println("Ok!")
}
