package service

import (
	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type StockPrice interface {
	Price(stk *stock.Stock) (stock.Price, error)
}

type StockPriceVolatility interface {
	PriceVolatility(stk *stock.Stock) (stock.PriceVolatility, error)
}

type StockSummary interface {
	Summary(stk *stock.Stock) (stock.Summary, error)
}

type StockDividend interface {
	NextFuture(stk *stock.Stock) (dividend.StockDividend, error)
	Future(stk *stock.Stock) ([]dividend.StockDividend, error)
	Historical(stk *stock.Stock, fromDate time.Time) ([]dividend.StockDividend, error)
}
