package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"os"
)

type DB struct {
	file      *os.File
	shortURLs map[string]*entity.ShortURL
}

type fileDTO struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func New(filePath string) (*DB, error) {
	var shortURLs = make(map[string]*entity.ShortURL)

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	err = restoreShortURLs(f, shortURLs)
	if err != nil {
		return nil, err
	}

	return &DB{
		file:      f,
		shortURLs: shortURLs,
	}, nil
}

func restoreShortURLs(f *os.File, shortURLs map[string]*entity.ShortURL) error {
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		dto := &fileDTO{}
		err := json.Unmarshal([]byte(scanner.Text()), dto)
		if err != nil {
			return fmt.Errorf(dbErrors.ErrRestoreFromFile.Error(), err.Error())
		}
		shortURL := toShortURL(dto)
		shortURLs[shortURL.Alias] = shortURL
	}

	return scanner.Err()
}

func toFileDTO(shortURL *entity.ShortURL) *fileDTO {
	return &fileDTO{
		UUID:        shortURL.UUID,
		ShortURL:    shortURL.Alias,
		OriginalURL: shortURL.SourceURL,
	}
}

func toShortURL(dto *fileDTO) *entity.ShortURL {
	return &entity.ShortURL{
		UUID:      dto.UUID,
		Alias:     dto.ShortURL,
		SourceURL: dto.OriginalURL,
	}
}

func (db *DB) Find(alias string) (*entity.ShortURL, error) {
	shortURL, ok := db.shortURLs[alias]
	if !ok {
		return nil, dbErrors.ErrNotFound
	}

	return shortURL, nil

}

func (db *DB) Save(shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	var err error
	var record *entity.ShortURL
	var data []byte

	if record, _ = db.Find(shortURL.Alias); record != nil {
		return nil, dbErrors.ErrIsNotUnique
	}

	db.shortURLs[shortURL.Alias] = shortURL

	data, err = json.Marshal(toFileDTO(shortURL))
	if err != nil {
		return nil, err
	}

	if _, err = db.file.WriteString(string(data) + "\n"); err != nil {
		return nil, err
	}

	return shortURL, nil
}
