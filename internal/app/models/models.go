package models

import (
	"github.com/gururuby/shortener/internal/app/utils"
)

type ShortURL struct {
	BaseURL string
	Alias   string
}

func NewShortURL(baseURL string) ShortURL {
	return ShortURL{
		Alias:   utils.GenerateRandomString(5),
		BaseURL: baseURL,
	}
}

func (s *ShortURL) AliasURL() string {
	return "http://localhost:8080/" + s.Alias
}
