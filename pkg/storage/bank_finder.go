package storage

import (
	"github.com/jmoiron/sqlx"

	"database/sql"
	"errors"
	"fmt"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
)

type (
	bankAccountFinder struct {
		db sqlx.Queryer
	}
)

var _ bank.Finder = &bankAccountFinder{}

func (f *bankAccountFinder) FindByAlias(alias string) (*bank.Account, error) {
	var b bank.Account

	query := `SELECT * FROM bank_account WHERE alias like $1`

	err := sqlx.Get(f.db, &b, query, alias)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.New(fmt.Sprintf("Select bank_account form alias %s", alias))
	}

	return &b, nil
}
