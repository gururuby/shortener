/*
Package generator provides utilities for generating unique identifiers.

It includes:
- UUID generation using google/uuid
- Custom alias generation with configurable length
- Random string generation with alphanumeric characters
- Error handling for invalid configurations
*/
package generator

import (
	"github.com/google/uuid"
	"github.com/gururuby/shortener/pkg/generator/errors"
	"math/rand"
	"time"
)

// Generator provides methods for generating unique identifiers.
// It can produce both UUIDs and custom aliases of specified length.
type Generator struct {
	aliasLength int // Length of generated aliases
}

// New creates a new Generator instance with the specified alias length.
// Parameters:
// - aliasLength: Desired length for generated aliases (must be positive)
// Returns:
// - *Generator: Initialized generator instance
func New(aliasLength int) *Generator {
	return &Generator{
		aliasLength: aliasLength,
	}
}

// Alias generates a random alphanumeric string of the configured length.
// Returns:
// - string: Generated alias
// - error: errors.ErrGeneratorEmptyAliasLength if length is invalid
func (g *Generator) Alias() (string, error) {
	return generateAlias(g.aliasLength)
}

// UUID generates a universally unique identifier (UUID v4).
// Returns:
// - string: Generated UUID in string format
func (g *Generator) UUID() string {
	return uuid.NewString()
}

// generateAlias creates a random alphanumeric string of specified length.
// Parameters:
// - length: Desired length of the alias
// Returns:
// - string: Generated alias
// - error: errors.ErrGeneratorEmptyAliasLength if length is invalid
func generateAlias(length int) (string, error) {
	if length < 1 {
		return "", errors.ErrGeneratorEmptyAliasLength
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b), nil
}
