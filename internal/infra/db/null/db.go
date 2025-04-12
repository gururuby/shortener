package db

type DB struct{}

func New() *DB {
	return &DB{}
}

func (db *DB) Find(alias string) (string, error) {
	return alias, nil
}

func (db *DB) Save(sourceURL string) (string, error) {
	return sourceURL, nil
}
