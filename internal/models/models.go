package models

import (
	"github.com/gururuby/shortener/internal/utils"
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

func (s *ShortURL) AliasURL(serverBaseURL string) string {
	return serverBaseURL + "/" + s.Alias

}
