package storage

import (
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
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
	query := `INSERT INTO stock(
				id,
				market_id,
				exchange_id,
				name,
				symbol,
				last_price_update,
				high_low_52_week_update,
				type,
				sector,
				industry
			  ) 
			  VALUES ($1, $2, $3, $4, upper($5), $6, $7, $8, $9, $10)`

	_, err := tx.Exec(
		query,
		s.ID,
		s.Market.ID,
		s.Exchange.ID,
		s.Name,
		s.Symbol,
		s.LastPriceUpdate,
		s.HighLow52WeekUpdate,
		s.Type.ID,
		s.Sector.ID,
		s.Industry.ID,
	)
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
	query := `UPDATE stock SET 
			  	value = $1,
			  	last_price_update = $2,
			  	change = $4,
			  	high_52_week = $5, 
			  	low_52_week = $6, 
			  	high_low_52_Week_update = $2,
			  	eps = $7,
			  	per = $8 
			  WHERE id = $3`

	_, err := tx.Exec(
		query,
		s.Value.Amount,
		time.Now(),
		s.ID,
		s.Change.Amount,
		s.High52Week.Amount,
		s.Low52Week.Amount,
		s.EPS,
		s.PER,
	)
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

func (p *stockPersister) UpdateHighLow52WeekPrice(s *stock.Stock) error {
	return transaction(p.db, func(tx *sqlx.Tx) error {
		if err := p.execUpdateHighLow52WeekPrice(tx, s); err != nil {
			return err
		}

		return nil
	})
}

func (p *stockPersister) execUpdateHighLow52WeekPrice(tx *sqlx.Tx, s *stock.Stock) error {
	query := `UPDATE stock SET high_52_week = $1, low_52_week = $2, high_low_52_Week_update = $3 WHERE id = $4`

	_, err := tx.Exec(query, s.High52Week.Amount, s.Low52Week.Amount, time.Now(), s.ID)
	if err != nil {
		return err
	}

	return nil
}
