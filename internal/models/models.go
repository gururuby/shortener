package models

import (
	"github.com/gururuby/shortener/internal/utils"
)

type ShortURL struct {
	Source string
	Alias  string
}

func NewShortURL(source string) ShortURL {
	return ShortURL{
		Alias:  utils.GenerateRandomString(5),
		Source: source,
	}
}

func (s *ShortURL) AliasURL(baseURL string) string {
	return baseURL + "/" + s.Alias

}
