package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	// walletItemFinder struct to hold necessary dependencies
	walletItemFinder struct {
		db sqlx.Queryer
	}
)

func NewWalletItemFinder(db sqlx.Queryer) *walletItemFinder {
	return &walletItemFinder{
		db: db,
	}
}

func (p *walletItemFinder) FindByStock(s *stock.Stock) (*wallet.Item, error) {
	return nil, mm.ErrNotFound
}
