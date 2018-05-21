package storage

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/jmoiron/sqlx"
)

type (
	// accountPersister struct to hold necessary dependencies
	accountPersister struct {
		db *sqlx.DB
	}
)

func NewAccountPersister(db *sqlx.DB) *accountPersister {
	return &accountPersister{
		db: db,
	}
}

func (p *accountPersister) PersistAll(as []*account.Account) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, a := range as {
			if err := p.execInsert(tx, a); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *accountPersister) execInsert(tx *sqlx.Tx, a *account.Account) error {
	query := `
		INSERT INTO account(
			id
			date, 
			stock_id, 
			operation, 
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
		a.ID,
		a.Date,
		a.Stock.ID,
		a.Operation,
		a.Amount,
		a.Price.Amount,
		a.PriceChange.Amount,
		a.PriceChangeCommission.Amount,
		a.Value.Amount,
		a.Commission.Amount,
	)
	if err != nil {
		return err
	}

	return nil
}
