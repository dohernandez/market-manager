package operation

import (
	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/satori/go.uuid"
)

type (
	Action string

	Operation struct {
		ID                    uuid.UUID
		Date                  time.Time
		Stock                 *stock.Stock
		Action                Action
		Amount                int
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
