package wallet

import (
	"github.com/satori/go.uuid"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type Item struct {
	ID       uuid.UUID
	Stock    *stock.Stock
	Amount   int
	Invested mm.Value
	Dividend mm.Value
	Buys     mm.Value
	Sells    mm.Value
}

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

func (i *Item) increaseInvestment(amount int, invested, priceChangeCommission, commission mm.Value) mm.Value {
	i.Amount = i.Amount + amount

	increase := invested.Increase(priceChangeCommission)
	increase = increase.Increase(commission)

	i.Invested = i.Invested.Increase(increase)
	i.Buys = i.Buys.Increase(increase)

	return increase
}

func (i *Item) decreaseInvestment(amount int, invested, priceChangeCommission, commission mm.Value) mm.Value {
	i.Amount = i.Amount - amount

	decrease := invested.Increase(priceChangeCommission)
	decrease = decrease.Increase(commission)

	i.Invested = i.Invested.Decrease(decrease)

	if i.Invested.Amount < 0 {
		i.Invested = mm.Value{}
	}

	i.Sells = i.Sells.Increase(invested)

	return decrease
}

func (i *Item) increaseDividend(dividend mm.Value) {
	i.Dividend = i.Dividend.Increase(dividend)
}

type Wallet struct {
	ID           uuid.UUID
	Name         string
	URL          string
	BankAccounts map[uuid.UUID]*bank.Account
	// stocks in trade
	Items map[uuid.UUID]*Item
	// inital capital
	Invested mm.Value
	// capital (sum of all stock values)
	Capital    mm.Value
	Funds      mm.Value
	Operations []*operation.Operation
}

func NewWallet(name, url string) *Wallet {
	return &Wallet{
		ID:           uuid.NewV4(),
		Name:         name,
		URL:          url,
		BankAccounts: map[uuid.UUID]*bank.Account{},
		Items:        map[uuid.UUID]*Item{},
	}
}

func (w *Wallet) AddBankAccount(ba *bank.Account) error {
	if _, ok := w.BankAccounts[ba.ID]; ok {
		return errors.New("account already exist")
	}

	w.BankAccounts[ba.ID] = ba

	return nil
}

func (w *Wallet) AddOperation(o *operation.Operation) {
	w.Operations = append(w.Operations, o)

	if o.Action == operation.Buy || o.Action == operation.Sell || o.Action == operation.Dividend {
		wi, ok := w.Items[o.Stock.ID]
		if !ok {
			wi = NewItem(o.Stock)
			w.Items[o.Stock.ID] = wi
		}

		switch o.Action {
		case operation.Buy:
			increased := wi.increaseInvestment(o.Amount, o.Value, o.PriceChangeCommission, o.Commission)
			w.decreaseFunds(increased)
		case operation.Sell:
			decreased := wi.decreaseInvestment(o.Amount, o.Value, o.PriceChangeCommission, o.Commission)
			w.increaseFunds(decreased)
		case operation.Dividend:
			wi.increaseDividend(o.Value)
			w.increaseFunds(o.Value)
		}
	}
}

func (w *Wallet) decreaseFunds(v mm.Value) {
	w.Funds = w.Funds.Decrease(v)
}

func (w *Wallet) increaseFunds(v mm.Value) {
	w.Funds = w.Funds.Increase(v)
}

func (w *Wallet) IncreaseInvestment(v mm.Value) {
	w.Invested = w.Invested.Increase(v)
	w.Capital = w.Capital.Increase(v)
	w.Funds = w.Funds.Increase(v)
}

func (w *Wallet) DecreaseInvestment(v mm.Value) {
	w.Invested = w.Invested.Decrease(v)
	w.Capital = w.Capital.Decrease(v)
	w.Funds = w.Funds.Decrease(v)
}
