package storage

import (
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	stockTuple struct {
		ID            uuid.UUID
		Name          string
		Symbol        string
		Value         string
		DividendYield string `db:"dividend_yield"`
		Change        string `db:"change"`

		MarketID          uuid.UUID `db:"market_id"`
		MarketName        string    `db:"market_name"`
		MarketDisplayName string    `db:"market_display_name"`

		ExchangeID     uuid.UUID `db:"exchange_id"`
		ExchangeName   string    `db:"exchange_name"`
		ExchangeSymbol string    `db:"exchange_symbol"`
	}

	stockFinder struct {
		db sqlx.Queryer
	}
)

func NewStockFinder(db sqlx.Queryer) *stockFinder {
	return &stockFinder{
		db: db,
	}
}

func (f *stockFinder) FindAll() ([]*stock.Stock, error) {
	var tuples []stockTuple

	query := `
		SELECT 
			s.id, s.name, s.symbol, s.market_id, s.exchange_id, s.value, s.dividend_yield, s.change,
			m.name AS market_name, m.display_name AS market_display_name,
			e.name AS exchange_name, e.symbol AS exchange_symbol
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
	`

	err := sqlx.Select(f.db, &tuples, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrap(err, "Select stocks")
	}

	var ss []*stock.Stock

	for _, s := range tuples {
		ss = append(ss, f.hydrate(&s))
	}

	return ss, nil
}

func (f *stockFinder) hydrate(s *stockTuple) *stock.Stock {
	dy, _ := strconv.ParseFloat(s.DividendYield, 64)

	return &stock.Stock{
		ID: s.ID,
		Market: &market.Market{
			ID:          s.MarketID,
			Name:        s.MarketName,
			DisplayName: s.MarketDisplayName,
		},
		Exchange: &exchange.Exchange{
			ID:     s.ExchangeID,
			Name:   s.ExchangeName,
			Symbol: s.ExchangeSymbol,
		},
		Name:          s.Name,
		Symbol:        s.Symbol,
		Value:         mm.ValueFromString(s.Value),
		DividendYield: dy,
		Change:        mm.ValueFromString(s.Change),
	}
}

func (f *stockFinder) FindBySymbol(symbol string) (*stock.Stock, error) {
	var tuple stockTuple

	query := `
		SELECT 
			s.id, s.name, s.symbol, s.market_id, s.exchange_id, s.value, s.dividend_yield, s.change,
			m.name AS market_name, m.display_name AS market_display_name,
			e.name AS exchange_name, e.symbol AS exchange_symbol
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
		WHERE s.symbol LIKE upper($1)
	`

	err := sqlx.Get(f.db, &tuple, query, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrap(err, "Select stock by symbol")
	}

	return f.hydrate(&tuple), nil
}

func (f *stockFinder) FindByName(name string) (*stock.Stock, error) {
	var tuple stockTuple

	query := `
		SELECT 
			s.id, s.name, s.symbol, s.market_id, s.exchange_id, s.value, s.dividend_yield, s.change,
			m.name AS market_name, m.display_name AS market_display_name,
			e.name AS exchange_name, e.symbol AS exchange_symbol
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
		WHERE s.name LIKE $1
	`

	err := sqlx.Get(f.db, &tuple, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrap(err, "Select stock by name")
	}

	return f.hydrate(&tuple), nil
}
