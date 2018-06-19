package wallet

import (
	"time"

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
	Operations  []operation.Operation
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
	capital := float64(i.Amount) * i.Stock.Value.Amount

	if i.CapitalRate > 0 {
		if i.Stock.Exchange.Symbol == "NASDAQ" || i.Stock.Exchange.Symbol == "NYSE" {
			capital = capital / i.CapitalRate
		}
	}

	return mm.Value{
		Amount:   capital,
		Currency: mm.Euro,
	}
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

func (i *Item) Change() mm.Value {
	change := float64(i.Amount) * i.Stock.Change.Amount

	if i.Stock.Exchange.Symbol == "NASDAQ" || i.Stock.Exchange.Symbol == "NYSE" {
		change = change / i.CapitalRate
	}

	return mm.Value{
		Amount:   change,
		Currency: mm.Euro,
	}
}

func (i *Item) WeightedAveragePrice() mm.Value {
	var asPrice float64

	eSymbol := i.Stock.Exchange.Symbol

	for _, o := range i.Operations {
		if o.Action != operation.Buy && o.Action != operation.Sell {
			continue
		}
		commissions := o.Commission.Increase(o.PriceChangeCommission)

		sPrice := o.Price.Amount * float64(o.Amount)

		if o.Action == operation.Buy {
			if mm.ExchangeCurrency(eSymbol) == mm.Dollar {
				sPrice = sPrice + commissions.Amount*o.PriceChange.Amount
			} else {
				sPrice = sPrice + commissions.Amount
			}

			asPrice = asPrice + sPrice
		} else {
			if mm.ExchangeCurrency(eSymbol) == mm.Dollar {
				sPrice = sPrice - commissions.Amount*o.PriceChange.Amount
			} else {
				sPrice = sPrice - commissions.Amount
			}

			asPrice = asPrice - sPrice
		}
	}

	wAveragePrice := asPrice / float64(i.Amount)
	return mm.Value{
		Amount:   wAveragePrice,
		Currency: mm.ExchangeCurrency(eSymbol),
	}
}

func (i *Item) PercentageInvestedRepresented(invested float64) float64 {
	return i.Invested.Amount * 100 / invested
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
	Dividend   mm.Value
	Commission mm.Value
	Connection mm.Value
	Interest   mm.Value

	// Rate currency conversion
	capitalRate float64
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
		w.Commission = w.Commission.Increase(o.FinalCommission())

	case operation.Sell:
		buyout := wi.decreaseInvestment(o.Amount, o.Value, o.PriceChangeCommission, o.Commission)

		w.Funds = w.Funds.Increase(buyout)
		w.Capital = w.Capital.Decrease(o.Capital())
		w.Commission = w.Commission.Increase(o.FinalCommission())

	case operation.Dividend:
		wi.increaseDividend(o.Value)

		w.Dividend = w.Dividend.Increase(o.Value)
		w.Funds = w.Funds.Increase(o.Value)

	case operation.Interest:
		w.Funds = w.Funds.Decrease(o.Value)
		w.Interest = w.Interest.Increase(o.Value)

	case operation.Connectivity:
		w.Funds = w.Funds.Decrease(o.Value)
		w.Connection = w.Connection.Increase(o.Value)
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

func (w *Wallet) NetBenefits() mm.Value {
	benefits := w.NetCapital()
	benefits = benefits.Decrease(w.Invested)

	return benefits
}

func (w *Wallet) NetCapital() mm.Value {
	netCapital := w.Capital.Increase(w.Funds)

	return netCapital
}

func (w *Wallet) PercentageBenefits() float64 {
	benefits := w.NetCapital()

	percent := (benefits.Amount * float64(100)) / w.Invested.Amount

	return percent - 100
}

func (w *Wallet) SetCapitalRate(capitalRate float64) {
	w.capitalRate = capitalRate

	for _, item := range w.Items {
		item.CapitalRate = capitalRate
	}
}

func (w *Wallet) DividendProjectedNextMonth() mm.Value {
	var dividends float64

	now := time.Now()
	month := now.Month()
	year := now.Year()

	for _, item := range w.Items {
		for _, d := range item.Stock.Dividends {
			if d.ExDate.Month() == month && d.ExDate.Year() == year {
				dividends = dividends + d.Amount.Amount*float64(item.Amount)
			}
		}
	}

	if w.capitalRate > 0 {
		dividends = dividends / w.capitalRate
	}

	return mm.Value{
		Amount:   dividends,
		Currency: mm.Euro,
	}
}

func (w *Wallet) Margin() mm.Value {
	netCapital := w.NetCapital()
	margin := netCapital.Amount * 49 / 100

	return mm.Value{
		Amount:   margin,
		Currency: netCapital.Currency,
	}
}

func (w *Wallet) FreeMargin() mm.Value {
	freeMargin := w.Margin()

	return freeMargin.Increase(w.Funds)
}

func (w *Wallet) DividendProjectedNextYear() mm.Value {
	var dividends float64

	now := time.Now()
	month := now.Month()
	year := now.Year()
	untilYear := now.Year() + 1

	for _, item := range w.Items {
		for _, d := range item.Stock.Dividends {
			if d.ExDate.Month() >= month && (d.ExDate.Year() >= year && d.ExDate.Year() < untilYear) {
				dividends = dividends + d.Amount.Amount*float64(item.Amount)
			}
		}
	}

	if w.capitalRate > 0 {
		dividends = dividends / w.capitalRate
	}

	return mm.Value{
		Amount:   dividends,
		Currency: mm.Euro,
	}
}
