package entity

import (
	"math/rand"
	"time"
)

const aliasLength = 5

type ShortURL struct {
	SourceURL string
	Alias     string
}

func NewShortURL(sourceURL string) ShortURL {
	return ShortURL{
		Alias:     generateAlias(aliasLength),
		SourceURL: sourceURL,
	}
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
