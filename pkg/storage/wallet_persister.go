package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type (
	// walletPersister struct to hold necessary dependencies
	walletPersister struct {
		db *sqlx.DB
	}
)

func NewWalletPersister(db *sqlx.DB) *walletPersister {
	return &walletPersister{
		db: db,
	}
}

func (p *walletPersister) PersistAll(ws []*wallet.Wallet) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, w := range ws {
			if err := p.execInsert(tx, w); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *walletPersister) execInsert(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `INSERT INTO wallet(id, name, url) VALUES ($1, $2, $3)`

	_, err := tx.Exec(query, w.ID, w.Name, w.URL)
	if err != nil {
		return err
	}

	return p.execBankAccountInsert(tx, w)
}

func (p *walletPersister) execBankAccountInsert(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `INSERT INTO wallet_bank_account(wallet_id, bank_acount_id) VALUES ($1, $2)`

	for _, ba := range w.BankAccounts {
		_, err := tx.Exec(query, w.ID, ba.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
