package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type (
	// transferPersister struct to hold necessary dependencies
	transferPersister struct {
		db *sqlx.DB
	}
)

var _ transfer.Persister = &transferPersister{}

func NewTransferPersister(db *sqlx.DB) *transferPersister {
	return &transferPersister{
		db: db,
	}
}

func (p *transferPersister) PersistAll(ts []*transfer.Transfer) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, t := range ts {
			if err := p.execInsert(tx, t); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *transferPersister) execInsert(tx *sqlx.Tx, t *transfer.Transfer) error {
	query := `
		INSERT INTO transfer(
			id, 
			from_account, 
			to_account, 
			amount, 
			date 
		) VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.Exec(
		query,
		t.ID,
		t.From.ID,
		t.To.ID,
		t.Amount.Amount,
		t.Date,
	)
	if err != nil {
		return err
	}

	return nil
}
