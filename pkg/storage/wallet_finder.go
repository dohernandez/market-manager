package storage

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	walletFinder struct {
		db sqlx.Queryer
	}

	walletTuple struct {
		ID       uuid.UUID `db:"id"`
		Name     string    `db:"name"`
		URL      string    `db:"url"`
		Invested string    `db:"invested"`
		Capital  string    `db:"capital"`
		Funds    string    `db:"funds"`
	}
)

var _ wallet.Finder = &walletFinder{}

func NewWalletFinder(db sqlx.Queryer) *walletFinder {
	return &walletFinder{
		db: db,
	}
}

func (f *walletFinder) FindByName(name string) (*wallet.Wallet, error) {
	var tuple walletTuple

	query := `SELECT * FROM wallet WHERE name like $1`

	err := sqlx.Get(f.db, &tuple, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrapf(err, "Select wallet with name %q", name)
	}

	return f.hydrateWallet(&tuple), nil
}

func (f *walletFinder) hydrateWallet(tuple *walletTuple) *wallet.Wallet {
	return &wallet.Wallet{
		ID:           tuple.ID,
		Name:         tuple.Name,
		URL:          tuple.URL,
		Invested:     mm.ValueFromString(tuple.Invested),
		Capital:      mm.ValueFromString(tuple.Capital),
		Funds:        mm.ValueFromString(tuple.Funds),
		BankAccounts: map[uuid.UUID]*bank.Account{},
		Items:        map[uuid.UUID]*wallet.Item{},
	}
}

func (f *walletFinder) FindByBankAccount(ba *bank.Account) (*wallet.Wallet, error) {
	var tuple walletTuple

	query := `SELECT w.* 
			FROM wallet w
			INNER JOIN wallet_bank_account wba ON w.ID = wba.wallet_id 
			WHERE wba.bank_account_id = $1`

	err := sqlx.Get(f.db, &tuple, query, ba.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrapf(err, "Select wallet with bank_account_id %q", ba.ID)
	}

	w := f.hydrateWallet(&tuple)

	err = w.AddBankAccount(ba)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (f *walletFinder) FindWalletsWithItemByStock(stk *stock.Stock) ([]*wallet.Wallet, error) {
	type walletWithWalletItemTuple struct {
		walletTuple
		ID     uuid.UUID `db:"wallet_item_id"`
		Amount int       `db:"wallet_item_amount"`
	}

	var tuples []walletWithWalletItemTuple

	query := `SELECT w.*,
			wi.ID as wallet_item_id, wi.amount as wallet_item_amount
			FROM wallet w
			INNER JOIN wallet_item wi ON w.ID = wi.wallet_id 
			WHERE wi.stock_id = $1`

	err := sqlx.Select(f.db, &tuples, query, stk.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrapf(err, "Select wallets with item with stock %q", stk.ID)
	}

	var ws []*wallet.Wallet
	for _, tuple := range tuples {
		w := f.hydrateWallet(&tuple.walletTuple)

		w.Items[stk.ID] = &wallet.Item{
			ID:     tuple.ID,
			Amount: tuple.Amount,
			Stock:  stk,
		}

		ws = append(ws, w)
	}

	return ws, nil
}
