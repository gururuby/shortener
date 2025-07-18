package validator

import "fmt"

func ExampleIsInvalidURL() {
	// Valid URL check
	fmt.Println(IsInvalidURL("https://example.com/path?query=param"))
	// Invalid URL check
	fmt.Println(IsInvalidURL("example.com"))

	// Output:
	// false
	// true
}
