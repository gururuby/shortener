package services

import (
	"io"
	"strings"
)

const (
	SomethingWentWrong = "Something went wrong"
	EmptySourceURL     = "Empty source URL, please specify URL"
	EmptyAliasError    = "Empty alias, please specify alias"
	SourceURLNotFound  = "Source URL not found"
)

type Storage interface {
	Save(string, string) (string, bool)
	Find(string) (string, bool)
}

func SaveURL(baseURL string, storage Storage, input io.ReadCloser) (result string, ok bool) {
	reqBody, err := io.ReadAll(input)
	if err != nil {
		return SomethingWentWrong, false
	}

	sourceURL := string(reqBody)

	if sourceURL == "" {
		return EmptySourceURL, false
	}

	return storage.Save(baseURL, sourceURL)
}

func FindURL(storage Storage, query string) (result string, ok bool) {
	alias := strings.TrimPrefix(query, "/")

	if alias == "" {
		return EmptyAliasError, false
	}

	sourceURL, ok := storage.Find(alias)

	if !ok {
		return SourceURLNotFound, false
	}

	return sourceURL, true
}
