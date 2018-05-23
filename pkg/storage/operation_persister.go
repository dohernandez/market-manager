package storage

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/jmoiron/sqlx"
)

type (
	// operationPersister struct to hold necessary dependencies
	operationPersister struct {
		db *sqlx.DB
	}
)

func NewOperationPersister(db *sqlx.DB) *operationPersister {
	return &operationPersister{
		db: db,
	}
}

func (p *operationPersister) PersistAll(os []*operation.Operation) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, o := range os {
			if err := p.execInsert(tx, o); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *operationPersister) execInsert(tx *sqlx.Tx, o *operation.Operation) error {
	query := `
		INSERT INTO operation(
			id,
			date, 
			stock_id, 
			action, 
			amount, 
			price, 
			price_change, 
			price_change_commission, 
			value, 
			commission
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := tx.Exec(
		query,
		o.ID,
		o.Date,
		o.Stock.ID,
		o.Action,
		o.Amount,
		o.Price.Amount,
		o.PriceChange.Amount,
		o.PriceChangeCommission.Amount,
		o.Value.Amount,
		o.Commission.Amount,
	)
	if err != nil {
		return err
	}

	return nil
}
