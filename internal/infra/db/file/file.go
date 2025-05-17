package db

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	shortURLEntity "github.com/gururuby/shortener/internal/domain/entity/shorturl"
	userEntity "github.com/gururuby/shortener/internal/domain/entity/user"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"os"
	"sync"
)

type FileDB struct {
	mutex     sync.RWMutex
	file      *os.File
	shortURLs map[string]*shortURLEntity.ShortURL
	users     map[int]*userEntity.User
}

type fileDTO struct {
	UserID      int    `json:"user_id"`
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func New(filePath string) (*FileDB, error) {
	var (
		shortURLs = make(map[string]*shortURLEntity.ShortURL)
		users     = make(map[int]*userEntity.User)
	)

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	err = restoreShortURLs(f, shortURLs)
	if err != nil {
		return nil, err
	}

	return &FileDB{
		file:      f,
		shortURLs: shortURLs,
		users:     users,
	}, nil
}

func restoreShortURLs(f *os.File, shortURLs map[string]*shortURLEntity.ShortURL) error {
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

func toFileDTO(shortURL *shortURLEntity.ShortURL) *fileDTO {
	return &fileDTO{
		UserID:      shortURL.UserID,
		UUID:        shortURL.UUID,
		ShortURL:    shortURL.Alias,
		OriginalURL: shortURL.SourceURL,
	}
}

func toShortURL(dto *fileDTO) *shortURLEntity.ShortURL {
	return &shortURLEntity.ShortURL{
		UserID:    dto.UserID,
		UUID:      dto.UUID,
		Alias:     dto.ShortURL,
		SourceURL: dto.OriginalURL,
	}
}

func (db *FileDB) FindUser(_ context.Context, id int) (*userEntity.User, error) {
	user, ok := db.users[id]
	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}
	return user, nil
}

func (db *FileDB) FindUserURLs(_ context.Context, userID int) ([]*shortURLEntity.ShortURL, error) {
	var urls []*shortURLEntity.ShortURL

	for _, url := range db.shortURLs {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

func (db *FileDB) SaveUser(_ context.Context) (*userEntity.User, error) {
	id := len(db.users) + 1
	user := &userEntity.User{ID: id}
	db.users[id] = user
	return user, nil
}

func (db *FileDB) FindShortURL(_ context.Context, alias string) (*shortURLEntity.ShortURL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	shortURL, ok := db.shortURLs[alias]

	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *FileDB) findShortURLBySourceURL(_ context.Context, sourceURL string) (*shortURLEntity.ShortURL, error) {
	var (
		shortURL  *shortURLEntity.ShortURL
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

func (db *FileDB) SaveShortURL(ctx context.Context, shortURL *shortURLEntity.ShortURL) (*shortURLEntity.ShortURL, error) {
	var (
		err    error
		record *shortURLEntity.ShortURL
		data   []byte
	)

	if record, _ = db.findShortURLBySourceURL(ctx, shortURL.SourceURL); record != nil {
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

func (db *FileDB) Ping(_ context.Context) error {
	_, err := db.file.Stat()
	return err
}
