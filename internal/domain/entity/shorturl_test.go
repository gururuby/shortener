package entity

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewShortURLOk(t *testing.T) {
	t.Run("create valid short URL entity", func(t *testing.T) {
		sourceURL := "https://google.com"
		got := NewShortURL(sourceURL)

		assert.Equal(t, got.SourceURL, sourceURL)
		assert.Regexp(t, "^\\w{5}$", got.Alias)
	})
}
