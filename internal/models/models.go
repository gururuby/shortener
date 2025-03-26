package models

import (
	"github.com/gururuby/shortener/internal/utils"
)

const AliasLength = 5

type ShortURL struct {
	Source string
	Alias  string
}

func NewShortURL(source string) ShortURL {
	return ShortURL{
		Alias:  utils.GenerateRandomString(AliasLength),
		Source: source,
	}
}

func (s *ShortURL) AliasURL(baseURL string) string {
	return baseURL + "/" + s.Alias

}
