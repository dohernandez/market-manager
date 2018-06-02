package storage

import (
	"github.com/jmoiron/sqlx"
	"github.com/satori/go.uuid"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	// stockDividendPersister struct to hold necessary dependencies
	stockDividendPersister struct {
		db *sqlx.DB
	}
)

func NewStockDividendPersister(db *sqlx.DB) *stockDividendPersister {
	return &stockDividendPersister{
		db: db,
	}
}

func (p *stockDividendPersister) PersistAll(stockID uuid.UUID, ds []dividend.StockDividend) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, d := range ds {
			if err := p.execInsert(tx, stockID, d); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *stockDividendPersister) execInsert(tx *sqlx.Tx, stockID uuid.UUID, d dividend.StockDividend) error {
	query := `
		INSERT INTO stock_dividend(
			stock_id, 
			ex_date, 
			payment_date, 
			record_date, 
			status, 
			amount, 
			change_from_prev, 
			change_from_prev_year, 
			prior_12_months_yield
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := tx.Exec(
		query,
		stockID,
		d.ExDate,
		d.PaymentDate,
		d.RecordDate,
		d.Status,
		d.Amount,
		d.ChangeFromPrev,
		d.ChangeFromPrevYear,
		d.Prior12MonthsYield,
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"INSERT INTO stock_dividend VALUE (%s, %s, %s, %s, %s, %f, %f, %f, %f)",
			stockID,
			d.ExDate,
			d.PaymentDate,
			d.RecordDate,
			d.Status,
			d.Amount,
			d.ChangeFromPrev,
			d.ChangeFromPrevYear,
			d.Prior12MonthsYield,
		)
	}

	return nil
}
