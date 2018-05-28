package wallet

import (
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type Item struct {
	ID          uuid.UUID
	Stock       *stock.Stock
	Amount      int
	Invested    mm.Value
	Dividend    mm.Value
	Buys        mm.Value
	Sells       mm.Value
	CapitalRate float64
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

	invested = invested.Increase(priceChangeCommission)
	invested = invested.Increase(commission)

	i.Invested = i.Invested.Increase(invested)
	i.Buys = i.Buys.Increase(invested)

	return invested
}

func (i *Item) decreaseInvestment(amount int, buyout, priceChangeCommission, commission mm.Value) mm.Value {
	i.Amount = i.Amount - amount

	buyout = buyout.Decrease(priceChangeCommission)
	buyout = buyout.Decrease(commission)

	i.Invested = i.Invested.Decrease(buyout)

	if i.Invested.Amount < 0 {
		i.Invested = mm.Value{}
	}

	i.Sells = i.Sells.Increase(buyout)

	return buyout
}

func (i *Item) increaseDividend(dividend mm.Value) {
	i.Dividend = i.Dividend.Increase(dividend)
}

func (i *Item) Capital() mm.Value {
	capital := float64(i.Amount) * i.Stock.Value.Amount / i.CapitalRate

	return mm.Value{Amount: capital}
}

func (i *Item) NetBenefits() mm.Value {
	benefits := i.benefits()
	benefits = benefits.Decrease(i.Buys)

	return benefits
}

func (i *Item) benefits() mm.Value {
	benefits := i.Capital()
	benefits = benefits.Increase(i.Sells)
	benefits = benefits.Increase(i.Dividend)

	return benefits
}

func (i *Item) PercentageBenefits() float64 {
	benefits := i.benefits()

	percent := (benefits.Amount * float64(100)) / i.Buys.Amount

	return percent - 100
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
	// capital (sum of all wallet item capital)
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

	wi, ok := w.Items[o.Stock.ID]
	if !ok {
		if o.Stock.ID != uuid.Nil {
			wi = NewItem(o.Stock)
			w.Items[o.Stock.ID] = wi
		}
	}

	switch o.Action {
	case operation.Buy:
		invested := wi.increaseInvestment(o.Amount, o.Value, o.PriceChangeCommission, o.Commission)

		w.Funds = w.Funds.Decrease(invested)
		w.Capital = w.Capital.Increase(o.Capital())

	case operation.Sell:
		buyout := wi.decreaseInvestment(o.Amount, o.Value, o.PriceChangeCommission, o.Commission)

		w.Funds = w.Funds.Increase(buyout)
		w.Capital = w.Capital.Decrease(o.Capital())

	case operation.Dividend:
		wi.increaseDividend(o.Value)

		w.Funds = w.Funds.Increase(o.Value)

	case operation.Interest:
		w.Funds = w.Funds.Decrease(o.Value)

	case operation.Connectivity:
		w.Funds = w.Funds.Decrease(o.Value)
	}
}

func (w *Wallet) IncreaseInvestment(v mm.Value) {
	w.Invested = w.Invested.Increase(v)
	w.Funds = w.Funds.Increase(v)
}

func (w *Wallet) DecreaseInvestment(v mm.Value) {
	w.Invested = w.Invested.Decrease(v)
	w.Funds = w.Funds.Decrease(v)
}

func (w *Wallet) UpdateCapital(v mm.Value) {
	w.Capital = w.Capital.Increase(v)
}
