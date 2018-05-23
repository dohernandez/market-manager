package wallet

import (
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	Item struct {
		ID       uuid.UUID
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
		ID:       uuid.NewV4(),
		Stock:    stock,
		Invested: mm.Value{},
		Dividend: mm.Value{},
		Buys:     mm.Value{},
		Sells:    mm.Value{},
	}
}

func (i *Item) IncreaseInvestment(amount int, invested, priceChangeCommission, commission mm.Value) {
	i.Amount = i.Amount + amount

	increased := invested.Increase(priceChangeCommission)
	increased = increased.Increase(commission)

	i.Invested = i.Invested.Increase(increased)
	i.Buys = i.Buys.Increase(increased)
}

func (i *Item) DecreaseInvestment(amount int, invested, priceChangeCommission, commission mm.Value) {
	i.Amount = i.Amount - amount

	decrease := invested.Increase(priceChangeCommission)
	decrease = decrease.Increase(commission)

	i.Invested = i.Invested.Decrease(decrease)

	if i.Invested.Amount < 0 {
		i.Invested = mm.Value{}
	}

	i.Sells = i.Sells.Increase(invested)
}

func (i *Item) IncreaseDividend(dividend mm.Value) {
	i.Dividend = i.Dividend.Increase(dividend)
}
