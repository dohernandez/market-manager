package account

import (
	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
	"github.com/satori/go.uuid"
)

type (
	Operation string

	Account struct {
		ID                    uuid.UUID
		Date                  time.Time
		Stock                 *stock.Stock
		Operation             Operation
		Amount                int
		Price                 mm.Value
		PriceChange           mm.Value
		PriceChangeCommission mm.Value
		Value                 mm.Value
		Commission            mm.Value
	}
)

const (
	Buy          Operation = "buy"
	Sell         Operation = "sell"
	Connectivity Operation = "connectivity"
	Dividend     Operation = "dividend"
	Interest     Operation = "interest"
)

func NewAccount(
	date time.Time,
	stock *stock.Stock,
	operation Operation,
	amount int,
	price,
	priceChange,
	priceChangeCommission,
	value,
	commission mm.Value,
) *Account {
	return &Account{
		ID:                    uuid.NewV4(),
		Date:                  date,
		Stock:                 stock,
		Operation:             operation,
		Amount:                amount,
		Price:                 price,
		PriceChange:           priceChange,
		PriceChangeCommission: priceChangeCommission,
		Value:      value,
		Commission: commission,
	}
}
