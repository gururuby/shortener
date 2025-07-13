/*
Package db implements a file-based database for the URL shortener service.

It provides:
- Persistent storage using JSON files
- In-memory caching for fast access
- Thread-safe operations with mutex locks
- Basic CRUD operations for users and short URLs
*/
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

// FileDB represents a file-based database implementation.
// It maintains in-memory maps synchronized with a persistent file.
type FileDB struct {
	mutex     sync.RWMutex                        // Mutex for thread-safe operations
	file      *os.File                            // Underlying file storage
	shortURLs map[string]*shortURLEntity.ShortURL // In-memory short URL cache
	users     map[int]*userEntity.User            // In-memory user cache
}

// fileDTO is the data transfer object for file storage.
// It defines the JSON structure for persisted short URLs.
type fileDTO struct {
	UserID      int    `json:"user_id"`      // Owner's user ID
	UUID        string `json:"uuid"`         // Unique identifier
	ShortURL    string `json:"short_url"`    // Short URL alias
	OriginalURL string `json:"original_url"` // Original long URL
	IsDeleted   bool   `json:"is_deleted"`   // Soft delete flag
}

// New creates and initializes a new FileDB instance.
// Parameters:
// - filePath: Path to the database file
// Returns:
// - *FileDB: Initialized database instance
// - error: If file operations fail
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

// restoreShortURLs loads existing short URLs from file into memory.
// Parameters:
// - f: File to read from
// - shortURLs: Map to populate with restored data
// Returns:
// - error: If reading or parsing fails
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

// toFileDTO converts a ShortURL entity to file storage format.
// Parameters:
// - shortURL: Entity to convert
// Returns:
// - *fileDTO: Data transfer object for storage
func toFileDTO(shortURL *shortURLEntity.ShortURL) *fileDTO {
	return &fileDTO{
		UserID:      shortURL.UserID,
		UUID:        shortURL.UUID,
		ShortURL:    shortURL.Alias,
		OriginalURL: shortURL.SourceURL,
		IsDeleted:   shortURL.IsDeleted,
	}
}

// toShortURL converts a fileDTO to ShortURL entity.
// Parameters:
// - dto: Data transfer object from storage
// Returns:
// - *shortURLEntity.ShortURL: Domain entity
func toShortURL(dto *fileDTO) *shortURLEntity.ShortURL {
	return &shortURLEntity.ShortURL{
		UserID:    dto.UserID,
		UUID:      dto.UUID,
		Alias:     dto.ShortURL,
		SourceURL: dto.OriginalURL,
		IsDeleted: dto.IsDeleted,
	}
}

// FindUser retrieves a user by ID.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - id: User ID to find
// Returns:
// - *userEntity.User: Found user
// - error: If user not found
func (db *FileDB) FindUser(_ context.Context, id int) (*userEntity.User, error) {
	user, ok := db.users[id]
	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}
	return user, nil
}

// FindUserURLs retrieves all short URLs belonging to a user.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - userID: Owner's user ID
// Returns:
// - []*shortURLEntity.ShortURL: List of user's URLs
// - error: Never returns error (empty slice for no results)
func (db *FileDB) FindUserURLs(_ context.Context, userID int) ([]*shortURLEntity.ShortURL, error) {
	var urls []*shortURLEntity.ShortURL

	for _, url := range db.shortURLs {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

// SaveUser creates and stores a new user.
// Parameters:
// - ctx: Context for cancellation/timeouts
// Returns:
// - *userEntity.User: Created user
// - error: Never returns error
func (db *FileDB) SaveUser(_ context.Context) (*userEntity.User, error) {
	id := len(db.users) + 1
	user := &userEntity.User{ID: id}
	db.users[id] = user
	return user, nil
}

// FindShortURL retrieves a short URL by its alias.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - alias: Short URL identifier
// Returns:
// - *shortURLEntity.ShortURL: Found short URL
// - error: If URL not found
func (db *FileDB) FindShortURL(_ context.Context, alias string) (*shortURLEntity.ShortURL, error) {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	shortURL, ok := db.shortURLs[alias]

	if !ok {
		return nil, dbErrors.ErrDBRecordNotFound
	}

	return shortURL, nil
}

// findShortURLBySourceURL looks up a short URL by its original URL.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - sourceURL: Original long URL
// Returns:
// - *shortURLEntity.ShortURL: Found short URL
// - error: If URL not found
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

// SaveShortURL stores a new short URL.
// Parameters:
// - ctx: Context for cancellation/timeouts
// - shortURL: URL to save
// Returns:
// - *shortURLEntity.ShortURL: Saved URL
// - error: If URL already exists or file operation fails
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

// MarkURLAsDeleted marks URLs as deleted (not implemented).
// Parameters:
// - ctx: Context for cancellation/timeouts
// - userID: Owner's user ID
// - aliases: URLs to mark as deleted
// Returns:
// - error: Always returns nil (not implemented)
func (db *FileDB) MarkURLAsDeleted(ctx context.Context, userID int, aliases []string) error {
	return nil
}

// Ping checks if the database is accessible.
// Parameters:
// - ctx: Context for cancellation/timeouts
// Returns:
// - error: If file stat operation fails
func (db *FileDB) Ping(_ context.Context) error {
	_, err := db.file.Stat()
	return err
}
