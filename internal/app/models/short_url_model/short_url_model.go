package short_url_model

import (
	"github.com/gururuby/shortener/internal/lib/utils/alias_generator"
)

type ShortUrl struct {
	BaseUrl string
	Alias   string
}

func (s *ShortUrl) GenerateAlias() {
	s.Alias = alias_generator.Run(5)
}

func (s *ShortUrl) AliasUrl() string {
	return "http://localhost:8080/" + s.Alias
}
