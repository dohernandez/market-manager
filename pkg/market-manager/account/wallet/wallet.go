package wallet

import (
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	Item struct {
		Stock    *stock.Stock
		Amount   int
		Invested mm.Value
		Dividend mm.Value
		Buys     mm.Value
		Sells    mm.Value
	}
)

func NewItem(stock *stock.Stock) *Item {
	return &Item{
		Stock:    stock,
		Invested: mm.Value{},
		Dividend: mm.Value{},
		Buys:     mm.Value{},
		Sells:    mm.Value{},
	}
}

func (i *Item) IncreaseInvestment(amount int, invested mm.Value) {
	i.Amount = i.Amount + amount
	i.Invested = i.Invested.Increase(invested)
	i.Buys = i.Buys.Increase(invested)
}

func (i *Item) DecreaseInvestment(amount int, invested mm.Value) {
	i.Amount = i.Amount - amount
	i.Invested = i.Invested.Decrease(invested)
	i.Sells = i.Sells.Increase(invested)
}

func (i *Item) IncreaseDividend(dividend mm.Value) {
	i.Dividend = i.Dividend.Increase(dividend)
}
