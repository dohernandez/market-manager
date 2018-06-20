package storage

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"database/sql"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	stockInfoFinder struct {
		db sqlx.Queryer
	}

	stockInfoTuple struct {
		ID   uuid.UUID `db:"id"`
		Type string    `db:"type"`
		Name string    `db:"name"`
	}
)

var _ stock.InfoFinder = &stockInfoFinder{}

func NewStockInfoFinder(db sqlx.Queryer) *stockInfoFinder {
	return &stockInfoFinder{
		db: db,
	}
}

func (f *stockInfoFinder) FindByName(name string) (*stock.Info, error) {
	var tuple stockInfoTuple

	query := `SELECT * FROM stock_info WHERE name like $1`

	err := sqlx.Get(f.db, &tuple, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.New(fmt.Sprintf("Select stock_info with name %q", name))
	}

	return &stock.Info{
		ID:   tuple.ID,
		Name: tuple.Name,
		Type: stock.InfoType(tuple.Type),
	}, nil
}
