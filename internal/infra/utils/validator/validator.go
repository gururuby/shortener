package validator

import "regexp"

func IsInvalidURL(rawURL string) bool {
	matched, _ := regexp.MatchString(`\Ahttps?://(www\.)?\w+(:\d+)?(\.\w+(:\d+)?)?.*\z`, rawURL)
	return !matched
}
