package storage

import (
	"github.com/jmoiron/sqlx"

	"fmt"

	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type WalletReload struct {
	db *sqlx.DB
}

func NewWalletReload(db *sqlx.DB) *WalletReload {
	return &WalletReload{
		db: db,
	}
}

func (rw *WalletReload) Reload(w *wallet.Wallet) error {
	return transaction(rw.db, func(tx *sqlx.Tx) error {
		if err := rw.execDeleteOperation(tx, w); err != nil {
			return err
		}

		if err := rw.execDeleteWalletItem(tx, w); err != nil {
			return err
		}

		if err := rw.execDeleteWalletItem(tx, w); err != nil {
			return err
		}

		if err := rw.execResetWallet(tx, w); err != nil {
			return err
		}

		if err := rw.execDeleteImport(tx, w); err != nil {
			return err
		}

		return nil
	})
}

func (rw *WalletReload) execDeleteOperation(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `DELETE FROM operation WHERE wallet_id = $1`

	_, err := tx.Exec(query, w.ID)
	if err != nil {
		return err
	}

	return nil
}

func (rw *WalletReload) execDeleteWalletItem(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `DELETE FROM wallet_item WHERE wallet_id = $1`

	_, err := tx.Exec(query, w.ID)
	if err != nil {
		return err
	}

	return nil
}

func (rw *WalletReload) execResetWallet(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `UPDATE wallet SET capital = 0, funds = invested, dividend = 0, commission = 0, connection = 0, interest = 0 WHERE id = $1`

	_, err := tx.Exec(query, w.ID)
	if err != nil {
		return err
	}

	return nil
}
func (rw *WalletReload) execDeleteImport(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `DELETE FROM import WHERE file_name LIKE $1`

	_, err := tx.Exec(query, fmt.Sprintf("%%_%s.%%", w.Name))
	if err != nil {
		return err
	}

	return nil
}
