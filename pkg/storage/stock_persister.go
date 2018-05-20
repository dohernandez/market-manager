package storage

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
	"github.com/jmoiron/sqlx"
)

type (
	// StockPersister struct to hold necessary dependencies
	stockPersister struct {
		db *sqlx.DB
	}
)

func NewStockPersister(db *sqlx.DB) *stockPersister {
	return &stockPersister{
		db: db,
	}
}

func (p *stockPersister) PersistAll(ss []*stock.Stock) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		for _, s := range ss {
			if err := p.execInsert(tx, s); err != nil {
				return err
			}
		}

		return nil
	})
}

func (p *stockPersister) execInsert(tx *sqlx.Tx, s *stock.Stock) error {
	query := `INSERT INTO stock(id, market_id, exchange_id, name, symbol) VALUES ($1, $2, $3, $4, upper($5))`

	_, err := tx.Exec(query, s.ID, s.Market.ID, s.Exchange.ID, s.Name, s.Symbol)
	if err != nil {
		return err
	}

	return nil
}

func (p *stockPersister) UpdatePrice(s *stock.Stock) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		if err := p.execUpdatePrice(tx, s); err != nil {
			return err
		}

		return nil
	})
}

func (p *stockPersister) execUpdatePrice(tx *sqlx.Tx, s *stock.Stock) error {
	query := `UPDATE stock SET value = $1 WHERE id = $2`

	_, err := tx.Exec(query, s.Value.Amount, s.ID)
	if err != nil {
		return err
	}

	return nil
}

func (p *stockPersister) UpdateDividendYield(s *stock.Stock) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		if err := p.execUpdateDividendYield(tx, s); err != nil {
			return err
		}

		return nil
	})
}

func (p *stockPersister) execUpdateDividendYield(tx *sqlx.Tx, s *stock.Stock) error {
	query := `UPDATE stock SET dividend_yield = $1 WHERE id = $2`

	_, err := tx.Exec(query, s.DividendYield, s.ID)
	if err != nil {
		return err
	}

	return nil
}
