package operation

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	Action string

	Operation struct {
		ID     uuid.UUID
		Date   time.Time
		Stock  *stock.Stock
		Action Action
		Amount int
		// Price in dollar
		Price                 mm.Value
		PriceChange           mm.Value
		PriceChangeCommission mm.Value
		Value                 mm.Value
		Commission            mm.Value
	}
)

const (
	Buy          Action = "buy"
	Sell         Action = "sell"
	Connectivity Action = "connectivity"
	Dividend     Action = "dividend"
	Interest     Action = "interest"
)

func NewOperation(
	date time.Time,
	stock *stock.Stock,
	action Action,
	amount int,
	price,
	priceChange,
	priceChangeCommission,
	value,
	commission mm.Value,
) *Operation {
	return &Operation{
		ID:                    uuid.NewV4(),
		Date:                  date,
		Stock:                 stock,
		Action:                action,
		Amount:                amount,
		Price:                 price,
		PriceChange:           priceChange,
		PriceChangeCommission: priceChangeCommission,
		Value:      value,
		Commission: commission,
	}
}

func (o *Operation) Capital() mm.Value {
	if o.Stock.ID == uuid.Nil {
		return mm.Value{}
	}

	capital := float64(o.Amount) * o.Stock.Value.Amount

	return mm.Value{Amount: capital}
}

func (o *Operation) FinalCommission() mm.Value {
	return o.Commission.Increase(o.PriceChangeCommission)
}

func (o *Operation) FinalPricePaid() mm.Value {
	fc := o.FinalCommission()

	p := o.Price
	if o.Price.Currency == mm.Dollar {
		p = mm.Value{
			Amount:   o.Price.Amount * o.PriceChange.Amount,
			Currency: mm.Euro,
		}
	}

	return fc.Increase(p)
}
