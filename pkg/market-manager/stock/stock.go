package stock

import (
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/market"
)

// Stock represents stock struct
type Stock struct {
	ID       uuid.UUID
	Market   *market.Market
	Exchange *exchange.Exchange
	Name     string
	Symbol   string
	Value    mm.Value
}

// NewStock creates an stock instance
func NewStock(market *market.Market, exchange *exchange.Exchange, name, symbol string) *Stock {
	return &Stock{
		ID:       uuid.NewV4(),
		Market:   market,
		Exchange: exchange,
		Name:     name,
		Symbol:   symbol,
	}
}
