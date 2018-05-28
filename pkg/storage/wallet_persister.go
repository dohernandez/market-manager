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

	err = p.execBankAccountInsert(tx, w)
	if err != nil {
		return err
	}

	return p.execOperationInsert(tx, w)
}

func (p *walletPersister) execBankAccountInsert(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `INSERT INTO wallet_bank_account(wallet_id, bank_account_id) VALUES ($1, $2)`

	for _, ba := range w.BankAccounts {
		_, err := tx.Exec(query, w.ID, ba.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *walletPersister) execOperationInsert(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `
		INSERT INTO operation(
			id,
			wallet_id,
			date, 
			stock_id, 
			action, 
			amount, 
			price, 
			price_change, 
			price_change_commission, 
			value, 
			commission
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	for _, o := range w.Operations {
		_, err := tx.Exec(
			query,
			o.ID,
			w.ID,
			o.Date,
			o.Stock.ID,
			o.Action,
			o.Amount,
			o.Price.Amount,
			o.PriceChange.Amount,
			o.PriceChangeCommission.Amount,
			o.Value.Amount,
			o.Commission.Amount,
		)
		if err != nil {
			return err
		}
	}

	return p.execWalletItemInsert(tx, w)
}

func (p *walletPersister) execWalletItemInsert(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `
		INSERT INTO wallet_item(id, wallet_id, stock_id, amount, invested, dividend, buys, sells) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET amount = excluded.amount, 
      		invested = excluded.invested,
      		dividend = excluded.dividend,
      		buys = excluded.buys,
      		sells = excluded.sells
	`

	for _, wi := range w.Items {
		_, err := tx.Exec(
			query,
			wi.ID,
			w.ID,
			wi.Stock.ID,
			wi.Amount,
			wi.Invested.Amount,
			wi.Dividend.Amount,
			wi.Buys.Amount,
			wi.Sells.Amount,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *walletPersister) PersistOperations(w *wallet.Wallet) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		err := p.execOperationInsert(tx, w)
		if err != nil {
			return err
		}

		return p.execUpdateFunds(tx, w)
	})
}

func (p *walletPersister) execUpdateFunds(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `UPDATE wallet SET funds = $1 WHERE id = $2`

	_, err := tx.Exec(query, w.Funds.Amount, w.ID)

	return err
}

func (p *walletPersister) UpdateAllAccounting(ws []*wallet.Wallet) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, w := range ws {
			if err := p.execUpdateAccounting(tx, w); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *walletPersister) execUpdateAccounting(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `UPDATE wallet SET funds = $1, capital = $2, invested = $3 WHERE id = $4`

	_, err := tx.Exec(query, w.Funds.Amount, w.Capital.Amount, w.Invested.Amount, w.ID)

	return err
}

func (p *walletPersister) UpdateAccounting(w *wallet.Wallet) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		return p.execUpdateAccounting(tx, w)
	})
}

// UpdateAllItemsCapital Update the capital of all items from all wallets, along with the capital of the wallet
func (p *walletPersister) UpdateAllItemsCapital(ws []*wallet.Wallet) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, w := range ws {
			if err := p.execUpdateItemCapital(tx, w); err != nil {
				return err
			}
			if err := p.execUpdateCapital(tx, w); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *walletPersister) execUpdateItemCapital(tx *sqlx.Tx, w *wallet.Wallet) error {
	for _, i := range w.Items {
		query := `UPDATE wallet_item SET capital = $1, capital_rate = $2 WHERE id = $3`

		_, err := tx.Exec(query, i.Capital().Amount, i.CapitalRate, i.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *walletPersister) execUpdateCapital(tx *sqlx.Tx, w *wallet.Wallet) error {
	query := `UPDATE wallet SET capital = (
		SELECT SUM(wi.capital)
		FROM wallet w
		INNER JOIN wallet_item wi ON w.id = wi.wallet_id
		WHERE w.id = $1
	) WHERE id = $2`

	_, err := tx.Exec(query, w.ID, w.ID)

	return err
}
