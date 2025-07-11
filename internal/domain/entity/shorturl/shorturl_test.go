package entity

import (
	"github.com/gururuby/shortener/internal/domain/entity/shorturl/mocks"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	"github.com/gururuby/shortener/pkg/generator/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func Test_NewShortURL_OK(t *testing.T) {
	t.Run("create valid short URL entity", func(t *testing.T) {
		sourceURL := "https://ya.ru"
		ctrl := gomock.NewController(t)
		generator := mocks.NewMockGenerator(ctrl)
		generator.EXPECT().UUID().Return("UUID").Times(1)
		generator.EXPECT().Alias().Return("alias", nil).Times(1)

		user := &userEntity.User{ID: 1}
		got, _ := NewShortURL(generator, user, sourceURL)

		assert.Equal(t, got.SourceURL, sourceURL)
		assert.Equal(t, got.UserID, 1)
		assert.Equal(t, got.IsDeleted, false)
		assert.Equal(t, "UUID", got.UUID)
		assert.Equal(t, "alias", got.Alias)
	})
}

func Test_NewShortURL_Errors(t *testing.T) {
	t.Run("when alias generating return error", func(t *testing.T) {
		sourceURL := "https://ya.ru"
		ctrl := gomock.NewController(t)
		generator := mocks.NewMockGenerator(ctrl)
		generator.EXPECT().Alias().Return("", errors.ErrGeneratorEmptyAliasLength).Times(1)

		user := &userEntity.User{ID: 1}
		_, err := NewShortURL(generator, user, sourceURL)

		require.Error(t, err)
	})
}
