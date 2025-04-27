//go:generate mockgen -destination=./mocks/mock.go -package=mocks . Client

package postgresql

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gururuby/shortener/config"
	"github.com/gururuby/shortener/internal/domain/entity"
	dbErrors "github.com/gururuby/shortener/internal/infra/db/errors"
	"github.com/gururuby/shortener/internal/infra/utils/retry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"sync"
)

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Ping(ctx context.Context) error
}

type DB struct {
	mutex     sync.RWMutex
	file      *os.File
	client    Client
	shortURLs map[string]*entity.ShortURL
}

type fileDTO struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func New(ctx context.Context, cfg *config.Config) (*DB, error) {
	var err error
	var f *os.File
	var client *pgxpool.Pool

	var shortURLs = make(map[string]*entity.ShortURL)

	f, err = os.OpenFile(cfg.FileStorage.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	err = restoreShortURLs(f, shortURLs)
	if err != nil {
		return nil, err
	}

	client, err = newClient(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	return &DB{
		file:      f,
		client:    client,
		shortURLs: make(map[string]*entity.ShortURL),
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

func newClient(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error
	var cancel context.CancelFunc

	err = utils.Retry(func() error {
		ctx, cancel = context.WithTimeout(ctx, cfg.ConnTryDelay)
		defer cancel()

		pool, err = pgxpool.New(ctx, cfg.DSN)

		if err != nil {
			return err
		}

		return nil

	}, cfg.ConnTryTimes, cfg.ConnTryDelay)

	return pool, err
}

func (db *DB) Find(alias string) (*entity.ShortURL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	shortURL, ok := db.shortURLs[alias]

	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

func (db *DB) Save(shortURL *entity.ShortURL) (*entity.ShortURL, error) {
	var err error
	var record *entity.ShortURL
	var data []byte

	if record, _ = db.Find(shortURL.Alias); record != nil {
		return nil, dbErrors.ErrDBIsNotUnique
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

func (db *DB) Ping() error {
	return db.client.Ping(context.Background())
}
