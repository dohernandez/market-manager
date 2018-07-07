package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	stockInfoPersister struct {
		db *sqlx.DB
	}
)

var _ stock.InfoPersister = &stockInfoPersister{}

func NewStockInfoPersister(db *sqlx.DB) *stockInfoPersister {
	return &stockInfoPersister{
		db: db,
	}
}

func (s *stockInfoPersister) Persist(i *stock.Info) error {
	return transaction(s.db, func(tx *sqlx.Tx) error {
		return s.execInsert(tx, i)
	})
}

func (s *stockInfoPersister) execInsert(tx *sqlx.Tx, i *stock.Info) error {
	query := `INSERT INTO stock_info(id, name, type) VALUES ($1, $2, $3)`

	_, err := tx.Exec(
		query,
		i.ID,
		i.Name,
		i.Type,
	)
	if err != nil {
		return err
	}

	return nil
}
