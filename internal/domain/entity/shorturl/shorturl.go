//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Generator

package entity

type Generator interface {
	UUID() string
	Alias() string
}

type ShortURL struct {
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

func NewShortURL(g Generator, sourceURL string) *ShortURL {
	return &ShortURL{
		UUID:      g.UUID(),
		Alias:     g.Alias(),
		SourceURL: sourceURL,
	}
}
