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

			if err := p.execBankAccountInsert(tx, w); err != nil {
				return err
			}

			if err := p.execOperationInsert(tx, w); err != nil {
				return err
			}

			if err := p.execWalletItemInsert(tx, w); err != nil {
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

	return nil
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

	return nil
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
		if err := p.execOperationInsert(tx, w); err != nil {
			return err
		}

		if err := p.execWalletItemInsert(tx, w); err != nil {
			return err
		}

		if err := p.execUpdateItemCapital(tx, w); err != nil {
			return err
		}

		if err := p.execUpdateCapital(tx, w); err != nil {
			return err
		}

		if err := p.execUpdateAccounting(tx, w); err != nil {
			return err
		}

		if err := p.execUpdateTrade(tx, w); err != nil {
			return err
		}

		return nil
	})
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
	query := `UPDATE wallet SET funds = $1, invested = $2, dividend = $3, commission = $4, connection = $5, interest = $6 WHERE id = $7`

	_, err := tx.Exec(
		query,
		w.Funds.Amount,
		w.Invested.Amount,
		w.Dividend.Amount,
		w.Commission.Amount,
		w.Connection.Amount,
		w.Interest.Amount,
		w.ID,
	)

	return err
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

func (p *walletPersister) execUpdateTrade(tx *sqlx.Tx, w *wallet.Wallet) error {
	queryTrade := `INSERT INTO trade(
				id,
				number,
				stock_id,
				wallet_id,
				opened_at,
				buys,
				buy_amount,
				amount,
				status,
				sells,
				sell_amount,
				dividend,
				closed_at,
				capital,
				net
			  ) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			  ON CONFLICT (id) DO UPDATE
			  SET amount = excluded.amount,
				  sells = excluded.sells,
				  sell_amount = excluded.sell_amount,
				  dividend = excluded.dividend,
				  closed_at = excluded.closed_at,
				  capital = excluded.capital,
				  net = excluded.net,
				  status = excluded.status`

	queryOperation := `INSERT INTO trade_operation(trade_id, operation_id) VALUES ($1, $2)
					   ON CONFLICT DO NOTHING`

	for _, t := range w.Trades {
		if _, err := tx.Exec(
			queryTrade,
			t.ID,
			t.Number,
			t.Stock.ID,
			w.ID,
			t.OpenedAt,
			t.Buys.Amount,
			t.BuyAmount,
			t.Amount,
			t.Status,
			t.Sells.Amount,
			t.SellAmount,
			t.Dividend.Amount,
			t.ClosedAt,
			t.CloseCapital.Amount,
			t.CloseNet.Amount,
		); err != nil {
			return err
		}

		for _, o := range t.Operations {
			if _, err := tx.Exec(queryOperation, t.ID, o.ID); err != nil {
				continue
			}
		}
	}

	return nil
}
