package storage

type IStorage interface {
	CreateShortURL(string, string) string
	FindShortURL(string) (string, bool)
}
