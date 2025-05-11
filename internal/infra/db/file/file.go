package file

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"os"
	"sync"
)

type DB struct {
	mutex     sync.RWMutex
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
			return fmt.Errorf(dbErrors.ErrDBRestoreFromFile.Error(), err.Error())
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

func (db *DB) Find(_ context.Context, alias string) (*entity.ShortURL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	shortURL, ok := db.shortURLs[alias]

	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *DB) findBySourceURL(_ context.Context, sourceURL string) (*entity.ShortURL, error) {
	var (
		shortURL  *entity.ShortURL
		noRecords = true
	)

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	for _, url := range db.shortURLs {
		if url.SourceURL == sourceURL {
			shortURL = url
			noRecords = false
			break
		}
	}

	if noRecords {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *DB) Save(ctx context.Context, shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	var err error
	var record *entity.ShortURL
	var data []byte

	if record, _ = db.findBySourceURL(ctx, shortURL.SourceURL); record != nil {
		return record, dbErrors.ErrDBIsNotUnique
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

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

func (db *DB) Ping(_ context.Context) error {
	_, err := db.file.Stat()
	return err
}
