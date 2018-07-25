package storage

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

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
		d.Amount.Amount,
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
			d.Amount.Amount,
			d.ChangeFromPrev,
			d.ChangeFromPrevYear,
			d.Prior12MonthsYield,
		)
	}

	return nil
}

func (p *stockDividendPersister) DeleteAllFromStatus(stockID uuid.UUID, status dividend.Status) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		return p.execDeleteFromStatus(tx, stockID, status)
	})
}

func (p *stockDividendPersister) execDeleteFromStatus(tx *sqlx.Tx, stockID uuid.UUID, status dividend.Status) error {
	query := `DELETE FROM stock_dividend WHERE stock_id = $1 AND status = $2`
	_, err := tx.Exec(query, stockID, status)
	if err != nil {
		return errors.Wrapf(err, "DELETE FROM stock_dividend WHERE stock_id = %s AND status = %s", stockID, status)
	}

	return nil
}

func (p *stockDividendPersister) DeleteAll(stockID uuid.UUID) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		return p.execDelete(tx, stockID)
	})
}

func (p *stockDividendPersister) execDelete(tx *sqlx.Tx, stockID uuid.UUID) error {
	query := `DELETE FROM stock_dividend WHERE stock_id = $1`
	_, err := tx.Exec(query, stockID)
	if err != nil {
		return errors.Wrapf(err, "DELETE FROM stock_dividend WHERE stock_id = %s", stockID)
	}

	return nil
}
