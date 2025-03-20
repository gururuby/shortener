package short_urls_repo

import (
	"github.com/gururuby/shortener/internal/app/models/short_url_model"
)

type ShortUrlsRepo struct {
	Data map[string]short_url_model.ShortUrl
}

func (r *ShortUrlsRepo) Create(baseUrl string) string {
	shortUrl := short_url_model.ShortUrl{BaseUrl: baseUrl}
	shortUrl.GenerateAlias()
	r.Data[shortUrl.Alias] = shortUrl

	return shortUrl.AliasUrl()
}

func (r *ShortUrlsRepo) Find(alias string) string {
	shortUrl := r.Data[alias]

	return shortUrl.BaseUrl
}
