package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/exchange"
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
	var m exchange.Exchange

	query := "SELECT * FROM exchange WHERE symbol=$1"

	err := sqlx.Get(f.db, &m, query, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return &m, mm.ErrNotFound
		}

		return &m, errors.Wrap(err, "Select exchange by symbol")
	}

	return &m, nil
}
