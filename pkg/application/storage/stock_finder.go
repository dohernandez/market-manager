package storage

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	stockTuple struct {
		ID                  uuid.UUID
		Name                string
		Symbol              string
		Value               string
		DividendYield       string    `db:"dividend_yield"`
		Change              string    `db:"change"`
		LastPriceUpdate     time.Time `db:"last_price_update"`
		High52week          string    `db:"high_52_week"`
		Low52week           string    `db:"low_52_week"`
		HighLow52WeekUpdate time.Time `db:"high_low_52_week_update"`
		EPS                 string    `db:"eps"`
		PER                 string    `db:"per"`

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

var _ stock.Finder = &stockFinder{}

func NewStockFinder(db sqlx.Queryer) *stockFinder {
	return &stockFinder{
		db: db,
	}
}

func (f *stockFinder) FindAll() ([]*stock.Stock, error) {
	var tuples []stockTuple

	query := fmt.Sprintf(`
		SELECT 
			%s
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
	`, strings.Join(f.selectStockColumns(), ", "))

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

func (f *stockFinder) selectStockColumns() []string {
	return []string{
		"s.id",
		"s.name",
		"s.symbol",
		"s.market_id",
		"s.exchange_id",
		"s.value",
		"s.dividend_yield",
		"s.change",
		"s.last_price_update",
		"s.high_52_week",
		"s.low_52_week",
		"s.high_low_52_week_update",
		"m.name AS market_name",
		"m.display_name AS market_display_name",
		"e.name AS exchange_name",
		"e.symbol AS exchange_symbol",
		"s.eps",
		"s.per",
	}
}

func (f *stockFinder) hydrate(tuple *stockTuple) *stock.Stock {
	dy, _ := strconv.ParseFloat(tuple.DividendYield, 64)
	eps, _ := strconv.ParseFloat(tuple.EPS, 64)
	per, _ := strconv.ParseFloat(tuple.PER, 64)

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
		Name:                tuple.Name,
		Symbol:              tuple.Symbol,
		Value:               mm.ValueFromStringAndExchange(tuple.Value, tuple.ExchangeSymbol),
		DividendYield:       dy,
		Change:              mm.ValueDollarFromString(tuple.Change),
		LastPriceUpdate:     tuple.LastPriceUpdate,
		High52Week:          mm.ValueFromStringAndExchange(tuple.High52week, tuple.ExchangeSymbol),
		Low52Week:           mm.ValueFromStringAndExchange(tuple.Low52week, tuple.ExchangeSymbol),
		HighLow52WeekUpdate: tuple.HighLow52WeekUpdate,
		EPS:                 eps,
		PER:                 per,
	}
}

func (f *stockFinder) FindBySymbol(symbol string) (*stock.Stock, error) {
	var tuple stockTuple

	query := fmt.Sprintf(`
		SELECT 
			%s
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
		WHERE s.symbol LIKE upper($1)
	`, strings.Join(f.selectStockColumns(), ", "))

	err := sqlx.Get(f.db, &tuple, query, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrapf(err, "Select stock by symbol %q", symbol)
	}

	return f.hydrate(&tuple), nil
}

func (f *stockFinder) FindByName(name string) (*stock.Stock, error) {
	var tuple stockTuple

	query := fmt.Sprintf(`
		SELECT 
			%s
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
		WHERE s.name LIKE $1
	`, strings.Join(f.selectStockColumns(), ", "))

	err := sqlx.Get(f.db, &tuple, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrapf(err, "FindByName %q", name)
	}

	return f.hydrate(&tuple), nil
}

func (f *stockFinder) FindByID(ID uuid.UUID) (*stock.Stock, error) {
	var tuple stockTuple

	query := fmt.Sprintf(`
		SELECT 
			%s
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
		WHERE s.id = $1
	`, strings.Join(f.selectStockColumns(), ", "))

	err := sqlx.Get(f.db, &tuple, query, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrapf(err, "Select stock by ID %q", ID)
	}

	return f.hydrate(&tuple), nil
}

func (f *stockFinder) FindAllByExchanges(exchanges []string) ([]*stock.Stock, error) {
	var tuples []stockTuple

	query := fmt.Sprintf(`
		SELECT 
			%s
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
		WHERE upper(e.symbol) IN (?)
	`, strings.Join(f.selectStockColumns(), ", "))

	query, args, err := sqlx.In(query, exchanges)
	if err != nil {
		return nil, err
	}

	err = sqlx.Select(f.db, &tuples, sqlx.Rebind(sqlx.DOLLAR, query), args...)
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

func (f *stockFinder) FindAllByDividendAnnounceProjectYearAndMonth(year, month int) ([]*stock.Stock, error) {
	type stockWithDividendsTuple struct {
		stockTuple
		DividendExDate     time.Time       `db:"dividend_ex_date"`
		DividendStatus     dividend.Status `db:"dividend_status"`
		DividendAmount     string          `db:"dividend_amount"`
		DividendRecordDate time.Time       `db:"dividend_record_date"`
	}

	var tuples []stockWithDividendsTuple

	query := fmt.Sprintf(`
		SELECT 
			%s,
           	sd.ex_date AS dividend_ex_date, sd.status AS dividend_status, sd.amount AS dividend_amount, sd.record_date AS dividend_record_date
		FROM stock s 
		INNER JOIN market m ON s.market_id = m.id
		INNER JOIN exchange e ON s.exchange_id = e.id
		INNER JOIN stock_dividend sd ON sd.stock_id = s.id
		WHERE sd.status IN ('announced', 'projected')
       	AND EXTRACT(YEAR FROM sd.ex_date) = $1
		AND EXTRACT(MONTH FROM sd.ex_date) = $2
	`, strings.Join(f.selectStockColumns(), ", "))
	err := sqlx.Select(f.db, &tuples, query, year, month)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mm.ErrNotFound
		}

		return nil, errors.Wrapf(
			err,
			"FindAllByDividendAnnounceProjectThisMonth with year: %q, month: %q",
			strconv.Itoa(year),
			strconv.Itoa(month),
		)
	}

	var ss []*stock.Stock

	for _, t := range tuples {
		s := f.hydrate(&t.stockTuple)
		s.Dividends = append(s.Dividends, dividend.StockDividend{
			ExDate:     t.DividendExDate,
			RecordDate: t.DividendRecordDate,
			Status:     t.DividendStatus,
			Amount:     mm.ValueDollarFromString(t.DividendAmount),
		})

		ss = append(ss, s)
	}

	return ss, nil
}
