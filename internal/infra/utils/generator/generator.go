package generator

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

type Generator struct {
	aliasLength int
}

func New(aliasLength int) *Generator {
	return &Generator{
		aliasLength: aliasLength,
	}
}

func (g *Generator) Alias() string {
	return generateAlias(g.aliasLength)
}

func (g *Generator) UUID() string {
	return uuid.NewString()
}

func generateAlias(length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b)
}
