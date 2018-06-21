package storage

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	stockDividendFinder struct {
		db sqlx.Queryer
	}

	dividendTuple struct {
		StockID            uuid.UUID `db:"stock_id"`
		ExDate             time.Time `db:"ex_date"`
		PaymentDate        time.Time `db:"payment_date"`
		RecordDate         time.Time `db:"record_date"`
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
	var tuples []dividendTuple

	query := "SELECT * FROM stock_dividend WHERE stock_id=$1 ORDER BY ex_date"

	err := sqlx.Select(f.db, &tuples, query, stockID)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Select dividend form stock id %s", stockID))
	}

	var ds []dividend.StockDividend
	for _, tuple := range tuples {
		ds = append(ds, f.hydrate(tuple))
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

	return f.hydrate(tuple), nil
}

func (f *stockDividendFinder) hydrate(tuple dividendTuple) dividend.StockDividend {
	cfp, _ := strconv.ParseFloat(tuple.ChangeFromPrev, 64)
	cfpy, _ := strconv.ParseFloat(tuple.ChangeFromPrevYear, 64)
	p12my, _ := strconv.ParseFloat(tuple.Prior12MonthsYield, 64)

	return dividend.StockDividend{
		ExDate:             tuple.ExDate,
		PaymentDate:        tuple.PaymentDate,
		RecordDate:         tuple.RecordDate,
		Status:             dividend.Status(tuple.Status),
		Amount:             mm.ValueDollarFromString(tuple.Amount),
		ChangeFromPrev:     cfp,
		ChangeFromPrevYear: cfpy,
		Prior12MonthsYield: p12my,
	}
}

func (f *stockDividendFinder) FindDividendNextAnnounceProjectFromYearAndMonth(ID uuid.UUID, year, month int) (dividend.StockDividend, error) {
	var tuple dividendTuple

	query := `
		SELECT *
		FROM stock_dividend sd
		WHERE sd.stock_id = $1 
		AND sd.status IN ('announced', 'projected')
       	AND EXTRACT(YEAR FROM sd.ex_date) >= $2
		AND EXTRACT(MONTH FROM sd.ex_date) >= $3
		ORDER BY sd.ex_date
	`

	err := sqlx.Get(f.db, &tuple, query, ID, year, month)
	if err != nil {
		if err == sql.ErrNoRows {
			return dividend.StockDividend{}, mm.ErrNotFound
		}

		return dividend.StockDividend{}, errors.Wrapf(
			err,
			"FindAllByDividendAnnounceProjectThisMonth with year: %q, month: %q",
			strconv.Itoa(year),
			strconv.Itoa(month),
		)
	}

	return f.hydrate(tuple), nil
}
