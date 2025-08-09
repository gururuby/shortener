// Package nomainpkg demonstrates a Go package with os.Exit usage that would be flagged
// by static analysis tools enforcing proper exit handling.
package nomainpkg

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("This is not main package")
	os.Exit(1)
}
