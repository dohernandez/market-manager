package storage

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type utilImportStorage struct {
	db *sqlx.DB
}

func NewUtilImportStorage(db *sqlx.DB) *utilImportStorage {
	return &utilImportStorage{
		db: db,
	}
}

func (s *utilImportStorage) Persist(r util.Resource) error {
	return transaction(s.db, func(tx *sqlx.Tx) error {
		return s.execInsert(tx, r)
	})
}

func (s *utilImportStorage) execInsert(tx *sqlx.Tx, r util.Resource) error {
	query := `
		INSERT INTO import(
			id, 
			resource, 
			file_name, 
			created_at 
		) VALUES ($1, $2, $3, $4)
	`

	_, err := tx.Exec(
		query,
		r.ID,
		r.Resource,
		r.FileName,
		r.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *utilImportStorage) FindAllByResource(resource string) ([]util.Resource, error) {
	var rs []util.Resource
	query := `SELECT * FROM import WHERE resource = $1`

	err := sqlx.Select(s.db, &rs, query, resource)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrap(err, fmt.Sprintf("Select import form resource %q", resource))
	}

	return rs, nil
}

func (s *utilImportStorage) FindLastByResourceAndWallet(resource, wName string) (util.Resource, error) {
	var r util.Resource
	query := `SELECT * FROM import WHERE resource = $1 AND file_name LIKE $2 ORDER BY created_at desc`

	err := sqlx.Get(s.db, &r, query, resource, fmt.Sprintf("%%_%s.%%", wName))
	if err != nil {
		if err == sql.ErrNoRows {
			return r, mm.ErrNotFound
		}

		return r, errors.Wrap(err, fmt.Sprintf("Select import form resource %q", resource))
	}

	return r, nil
}
