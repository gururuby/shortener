package entity

//go:generate mockgen -destination=./mock_entity/mock.go . Generator
type Generator interface {
	UUID() string
	Alias() string
}

type ShortURL struct {
	UUID      string
	SourceURL string
	Alias     string
}

func NewShortURL(g Generator, sourceURL string) *ShortURL {
	return &ShortURL{
		UUID:      g.UUID(),
		Alias:     g.Alias(),
		SourceURL: sourceURL,
	}
}
