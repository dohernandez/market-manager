package storage

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/import"
	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type importStorage struct {
	db *sqlx.DB
}

func NewImportStorage(db *sqlx.DB) *importStorage {
	return &importStorage{
		db: db,
	}
}

func (s *importStorage) Persist(r _import.Resource) error {
	return transaction(s.db, func(tx *sqlx.Tx) error {
		return s.execInsert(tx, r)
	})
}

func (s *importStorage) execInsert(tx *sqlx.Tx, r _import.Resource) error {
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

func (s *importStorage) FindAllByResource(resource string) ([]_import.Resource, error) {
	var rs []_import.Resource
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
