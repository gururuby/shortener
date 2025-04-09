package dao

import "errors"

const (
	maxGenerationAttempts = 5
)

var (
	errNonUnique = errors.New("record already exist")
	errSave      = errors.New("cannot save error")
)

type shortURLDB interface {
	Find(string) (string, error)
	Save(string) (string, error)
}

type ShortURLDAO struct {
	db shortURLDB
}

func NewShortURLDAO(db shortURLDB) *ShortURLDAO {
	return &ShortURLDAO{
		db: db,
	}
}

func (dao *ShortURLDAO) FindByAlias(alias string) (string, error) {
	return dao.db.Find(alias)
}

func (dao *ShortURLDAO) Save(sourceURL string) (string, error) {
	return dao.saveWithAttempt(1, sourceURL)
}

func (dao *ShortURLDAO) saveWithAttempt(attCount int, sourceURL string) (string, error) {
	if attCount > maxGenerationAttempts {
		return "", errSave
	}

	res, err := dao.db.Save(sourceURL)

	if errors.Is(err, errNonUnique) {
		attCount++
		return dao.saveWithAttempt(attCount, sourceURL)
	}

	return res, err
}
