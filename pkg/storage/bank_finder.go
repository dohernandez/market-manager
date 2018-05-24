package storage

import (
	"errors"
	"fmt"

	"github.com/almerlucke/go-iban/iban"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	"database/sql"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
)

type (
	bankAccountFinder struct {
		db sqlx.Queryer
	}

	bankAccountTuple struct {
		ID            uuid.UUID `db:"id"`
		Name          string    `db:"name"`
		AccountNo     string    `db:"account_no"`
		Alias         string    `db:"alias"`
		AccountNoType string    `db:"account_no_type"`
	}
)

var _ bank.Finder = &bankAccountFinder{}

func NewBankAccountFinder(db sqlx.Queryer) *bankAccountFinder {
	return &bankAccountFinder{
		db: db,
	}
}

func (f *bankAccountFinder) FindByAlias(alias string) (*bank.Account, error) {
	var tuple bankAccountTuple

	query := `SELECT * FROM bank_account WHERE alias like $1`

	err := sqlx.Get(f.db, &tuple, query, alias)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.New(fmt.Sprintf("Select bank_account form alias %q", alias))
	}

	var AccountNo string
	if bank.AccountNoType(tuple.AccountNoType) == bank.IBAN {
		IBAN, err := iban.NewIBAN(tuple.AccountNo)
		if err != nil {
			return nil, err
		}

		AccountNo = IBAN.PrintCode
	}

	return &bank.Account{
		ID:        tuple.ID,
		Name:      tuple.Name,
		AccountNo: AccountNo,
		Alias:     alias,
	}, nil
}
