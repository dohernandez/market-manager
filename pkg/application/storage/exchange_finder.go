package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
)

type exchangeFinder struct {
	db sqlx.Queryer
}

func NewExchangeFinder(db sqlx.Queryer) *exchangeFinder {
	return &exchangeFinder{
		db: db,
	}
}

func (f *exchangeFinder) FindBySymbol(symbol string) (*exchange.Exchange, error) {
	var e exchange.Exchange

	query := "SELECT * FROM exchange WHERE symbol=$1"

	err := sqlx.Get(f.db, &e, query, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrap(err, "Select exchange by symbol")
	}

	return &e, nil
}
