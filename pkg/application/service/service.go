package service

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
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
