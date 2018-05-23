package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type (
	// walletItemPersister struct to hold necessary dependencies
	walletItemPersister struct {
		db *sqlx.DB
	}
)

func NewWalletItemPersister(db *sqlx.DB) *walletItemPersister {
	return &walletItemPersister{
		db: db,
	}
}

func (p *walletItemPersister) PersistAll(wis []*wallet.Item) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, wi := range wis {
			if err := p.execInsert(tx, wi); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *walletItemPersister) execInsert(tx *sqlx.Tx, wi *wallet.Item) error {
	query := `INSERT INTO wallet_item(id, stock_id, amount, invested, dividend, buys, sells) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.Exec(query, wi.ID, wi.Stock.ID, wi.Amount, wi.Invested.Amount, wi.Dividend.Amount, wi.Buys.Amount, wi.Sells.Amount)
	if err != nil {
		return err
	}

	return nil
}
