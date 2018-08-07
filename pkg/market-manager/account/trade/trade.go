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

		OpenedAt   time.Time
		Buys       mm.Value
		BuysAmount float64

		ClosedAt    time.Time
		Sells       mm.Value
		SellsAmount float64

		Amount   float64
		Dividend mm.Value

		Status Status

		CloseCapital mm.Value
		CloseNet     mm.Value

		Stock      *stock.Stock
		Operations []*operation.Operation
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
	amount := float64(op.Amount)

	t.OpenedAt = op.Date
	t.BuysAmount = amount
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
		Amount: t.Stock.Value.Amount * t.Amount,
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

func (t *Trade) Sold(op *operation.Operation) {
	t.Operations = append(t.Operations, op)

	t.Sells = t.Sells.Increase(op.FinalPricePaid())

	t.SellsAmount += float64(op.Amount)
	t.Amount = t.BuysAmount - t.SellsAmount

	if t.Amount == 0 {
		t.closeTrade(op.Date)
	}
}

func (t *Trade) Close(op *operation.Operation) {
	t.Operations = append(t.Operations, op)

	t.Sells = t.Sells.Increase(op.FinalPricePaid())
	t.SellsAmount += float64(op.Amount)
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
