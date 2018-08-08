package trade

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	Status string

	Trade struct {
		ID     uuid.UUID
		Number int

		OpenedAt  time.Time
		Buys      mm.Value
		BuyAmount float64

		ClosedAt   time.Time
		Sells      mm.Value
		SellAmount float64

		Amount   float64
		Dividend mm.Value

		Status Status

		CloseCapital mm.Value
		CloseNet     mm.Value

		Stock      *stock.Stock
		Operations []*operation.Operation

		// Rate currency conversion
		CapitalRate float64
	}
)

const (
	Open    Status = "open"
	Close   Status = "close"
	Initial Status = "initial"
)

func NewTrade(number int) *Trade {
	return &Trade{
		ID:     uuid.NewV4(),
		Number: number,
		Status: Initial,
	}
}

func (t *Trade) Open(op *operation.Operation) {
	t.Operations = append(t.Operations, op)

	amount := float64(op.Amount)

	t.OpenedAt = op.Date
	t.BuyAmount = amount
	t.Amount = amount
	t.Buys = op.FinalPricePaid()
	t.Status = Open
	t.Stock = op.Stock
}

func (t *Trade) Capital() mm.Value {
	if t.Status == Close {
		return t.CloseCapital
	}

	capital := mm.Value{
		Amount:   t.Stock.Value.Amount * t.Amount / t.CapitalRate,
		Currency: mm.Euro,
	}

	return capital
}

func (t *Trade) Net() mm.Value {
	if t.Status == Close {
		return t.CloseNet
	}

	net := t.Sells.Increase(t.Capital())
	net = net.Increase(t.Dividend)
	net = net.Decrease(t.Buys)

	return net
}

func (t *Trade) BenefitPercentage() float64 {
	net := t.Net()

	return net.Amount * 100 / t.Buys.Amount
}

func (t *Trade) Sold(op *operation.Operation) {
	t.Operations = append(t.Operations, op)

	t.Sells = t.Sells.Increase(op.FinalPricePaid())

	t.SellAmount += float64(op.Amount)
	t.Amount = t.BuyAmount - t.SellAmount

	if t.Amount == 0 {
		t.closeTrade(op.Date)
	}
}

func (t *Trade) Bought(op *operation.Operation) {
	t.Operations = append(t.Operations, op)

	t.Buys = t.Buys.Increase(op.FinalPricePaid())

	t.BuyAmount += float64(op.Amount)
	t.Amount = t.BuyAmount - t.SellAmount
}

func (t *Trade) Close(op *operation.Operation) {
	t.Operations = append(t.Operations, op)

	t.Sells = t.Sells.Increase(op.FinalPricePaid())
	t.SellAmount += float64(op.Amount)
	t.Amount = 0

	t.closeTrade(op.Date)
}

func (t *Trade) closeTrade(closeAt time.Time) {
	t.Status = Close
	t.ClosedAt = closeAt

	t.CloseCapital = mm.Value{Currency: mm.Euro}

	net := t.Sells.Increase(t.Dividend)
	t.CloseNet = net.Decrease(t.Buys)
}

func (t *Trade) PayedDividend(d mm.Value) {
	t.Dividend = t.Dividend.Increase(d)
}

func (t *Trade) WeightedAverageBuyPrice() mm.Value {
	return t.weightedAveragePrice(operation.Buy)
}

func (t *Trade) weightedAveragePrice(action operation.Action) mm.Value {
	var asPrice float64

	eSymbol := t.Stock.Exchange.Symbol

	for _, o := range t.Operations {
		if o.Action != action {
			continue
		}

		commissions := o.Commission.Increase(o.PriceChangeCommission)

		sPrice := o.Price.Amount * float64(o.Amount)

		if mm.ExchangeCurrency(eSymbol) == mm.Dollar {
			sPrice = sPrice + commissions.Amount*o.PriceChange.Amount
		} else {
			sPrice = sPrice + commissions.Amount
		}

		asPrice = asPrice + sPrice
	}

	wAPrice := mm.Value{
		Currency: mm.ExchangeCurrency(eSymbol),
	}

	if t.BuyAmount > 0 {
		wAPrice.Amount = asPrice / float64(t.BuyAmount)
	}

	return wAPrice
}

func (t *Trade) WeightedAverageSellPrice() mm.Value {
	return t.weightedAveragePrice(operation.Sell)
}
