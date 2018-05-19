package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/market"
)

type marketFinder struct {
	db sqlx.Queryer
}

func NewMarketFinder(db sqlx.Queryer) *marketFinder {
	return &marketFinder{
		db: db,
	}
}

func (f *marketFinder) FindByName(name string) (*market.Market, error) {
	var m market.Market

	query := "SELECT * FROM market WHERE name=$1"

	err := sqlx.Get(f.db, &m, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrap(err, "Select market by name")
	}

	return &m, nil
}
