package storage

import (
	"database/sql"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
)

type (
	stockTuple struct {
		ID            uuid.UUID
		Name          string
		Symbol        string
		Value         string
		DividendYield string `db:"dividend_yield"`

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
			s.id, s.name, s.symbol, s.market_id, s.exchange_id, s.value, s.dividend_yield,
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
		dy, _ := strconv.ParseFloat(s.DividendYield, 64)
		v, _ := strconv.ParseFloat(s.Value, 64)

		ss = append(ss, &stock.Stock{
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
			Value:         mm.Value{Amount: v},
			DividendYield: dy,
		})
	}

	return ss, nil
}

func (f *stockFinder) FindBySymbol(symbol string) (*stock.Stock, error) {
	var tuple stockTuple

	query := `
		SELECT 
			s.id, s.name, s.symbol, s.market_id, s.exchange_id, s.value, s.dividend_yield,
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

	dy, _ := strconv.ParseFloat(tuple.DividendYield, 64)
	v, _ := strconv.ParseFloat(tuple.Value, 64)

	return &stock.Stock{
		ID: tuple.ID,
		Market: &market.Market{
			ID:          tuple.MarketID,
			Name:        tuple.MarketName,
			DisplayName: tuple.MarketDisplayName,
		},
		Exchange: &exchange.Exchange{
			ID:     tuple.ExchangeID,
			Name:   tuple.ExchangeName,
			Symbol: tuple.ExchangeSymbol,
		},
		Name:          tuple.Name,
		Symbol:        tuple.Symbol,
		Value:         mm.Value{Amount: v},
		DividendYield: dy,
	}, nil
}
