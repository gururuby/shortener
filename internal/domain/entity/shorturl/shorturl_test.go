package entity

import (
	"github.com/gururuby/shortener/internal/domain/entity/shorturl/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_NewShortURL_OK(t *testing.T) {
	t.Run("create valid short URL entity", func(t *testing.T) {
		sourceURL := "https://ya.ru"
		ctrl := gomock.NewController(t)
		generator := mocks.NewMockGenerator(ctrl)
		generator.EXPECT().UUID().Return("UUID").Times(1)
		generator.EXPECT().Alias().Return("alias").Times(1)

		got := NewShortURL(generator, sourceURL)

		assert.Equal(t, got.SourceURL, sourceURL)
		assert.Equal(t, "UUID", got.UUID)
		assert.Equal(t, "alias", got.Alias)
	})
}
