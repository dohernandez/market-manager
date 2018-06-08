package storage

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"fmt"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	stockDividendFinder struct {
		db sqlx.Queryer
	}

	dividendTuple struct {
		StockID            uuid.UUID `db:"stock_id"`
		ExDate             string    `db:"ex_date"`
		PaymentDate        string    `db:"payment_date"`
		RecordDate         string    `db:"record_date"`
		Status             string    `db:"status"`
		Amount             string    `db:"amount"`
		ChangeFromPrev     string    `db:"change_from_prev"`
		ChangeFromPrevYear string    `db:"change_from_prev_year"`
		Prior12MonthsYield string    `db:"prior_12_months_yield"`
	}
)

func NewStockDividendFinder(db sqlx.Queryer) *stockDividendFinder {
	return &stockDividendFinder{
		db: db,
	}
}

func (f *stockDividendFinder) FindAllFormStock(stockID uuid.UUID) ([]dividend.StockDividend, error) {
	var ds []dividend.StockDividend

	query := "SELECT * FROM stock_dividend WHERE stock_id=$1"

	err := sqlx.Get(f.db, &ds, query, stockID)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Select dividend form stock id %s", stockID))
	}

	return ds, nil
}

// FindNextFromStock Find the next dividend announce or projected for the stock
func (f *stockDividendFinder) FindNextFromStock(stockID uuid.UUID, dt time.Time) (dividend.StockDividend, error) {
	var tuple dividendTuple

	query := `
		SELECT *
		FROM stock_dividend
		WHERE stock_id = $1
		AND ex_date >= $2 
	`

	err := sqlx.Get(f.db, &tuple, query, stockID, dt)
	if err != nil {
		if err == sql.ErrNoRows {
			return dividend.StockDividend{}, mm.ErrNotFound
		}

		return dividend.StockDividend{}, errors.Wrap(err, "Select stock by symbol")
	}

	cfp, _ := strconv.ParseFloat(tuple.ChangeFromPrev, 64)
	cfpy, _ := strconv.ParseFloat(tuple.ChangeFromPrevYear, 64)
	p12my, _ := strconv.ParseFloat(tuple.Prior12MonthsYield, 64)

	return dividend.StockDividend{
		ExDate:             parseDateString(tuple.ExDate),
		PaymentDate:        parseDateString(tuple.PaymentDate),
		RecordDate:         parseDateString(tuple.RecordDate),
		Status:             dividend.Status(tuple.Status),
		Amount:             mm.ValueFromString(tuple.Amount),
		ChangeFromPrev:     cfp,
		ChangeFromPrevYear: cfpy,
		Prior12MonthsYield: p12my,
	}, nil
}
