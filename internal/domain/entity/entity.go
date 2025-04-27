package entity

//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Generator

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
