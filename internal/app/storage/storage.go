package storage

type StorageInterface interface {
	CreateShortURL(string) string
	FindShortURL(string) (string, bool)
}
