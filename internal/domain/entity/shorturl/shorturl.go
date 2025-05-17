//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Generator

package entity

import userEntity "github.com/gururuby/shortener/internal/domain/entity/user"

type Generator interface {
	UUID() string
	Alias() string
}

type ShortURL struct {
	UserID    int
	UUID      string
	SourceURL string
	Alias     string
}

type BatchShortURLInput struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortURLOutput struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewShortURL(g Generator, user *userEntity.User, sourceURL string) *ShortURL {
	shortURL := &ShortURL{
		UUID:      g.UUID(),
		Alias:     g.Alias(),
		SourceURL: sourceURL,
	}

	if user != nil {
		shortURL.UserID = user.ID
	}
	return shortURL
}
