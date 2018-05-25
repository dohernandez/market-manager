package stock

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	// Stock represents stock struct
	Stock struct {
		ID            uuid.UUID
		Market        *market.Market
		Exchange      *exchange.Exchange
		Name          string
		Symbol        string
		Value         mm.Value
		Dividends     []dividend.StockDividend
		DividendYield float64
		Change        mm.Value
	}

	// Price represents stock's price struct
	Price struct {
		Date   time.Time `json:"date"`
		Close  float64   `json:"close"`
		High   float64   `json:"high"`
		Low    float64   `json:"low"`
		Open   float64   `json:"open"`
		Volume float64   `json:"volume"`
	}
)

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
