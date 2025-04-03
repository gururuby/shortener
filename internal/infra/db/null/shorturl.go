package null

type ShortURLDB struct{}

func NewShortURLDB() *ShortURLDB {
	return &ShortURLDB{}
}

func (db *ShortURLDB) Find(alias string) (string, error) {
	return alias, nil
}

func (db *ShortURLDB) Save(sourceURL string) (string, error) {
	return sourceURL, nil
}
