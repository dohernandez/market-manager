package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"fmt"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock/dividend"
)

type stockDividendFinder struct {
	db sqlx.Queryer
}

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
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrap(err, fmt.Sprintf("Select dividend form stock id %s", stockID))
	}

	return ds, nil
}
