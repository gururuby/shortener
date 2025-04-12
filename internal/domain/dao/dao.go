package dao

import "errors"

const (
	maxGenerationAttempts = 5
)

var (
	errNonUnique = errors.New("record already exist")
	errSave      = errors.New("cannot save error")
)

type DB interface {
	Find(string) (string, error)
	Save(string) (string, error)
}

type DAO struct {
	db DB
}

func New(db DB) *DAO {
	return &DAO{
		db: db,
	}
}

func (dao *DAO) FindByAlias(alias string) (string, error) {
	return dao.db.Find(alias)
}

func (dao *DAO) Save(sourceURL string) (string, error) {
	return dao.saveWithAttempt(1, sourceURL)
}

func (dao *DAO) saveWithAttempt(attCount int, sourceURL string) (string, error) {
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
