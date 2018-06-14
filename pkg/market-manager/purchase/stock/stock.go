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
		ID                  uuid.UUID
		Market              *market.Market
		Exchange            *exchange.Exchange
		Name                string
		Symbol              string
		Value               mm.Value
		Dividends           []dividend.StockDividend
		DividendYield       float64
		Change              mm.Value
		LastPriceUpdate     time.Time
		High52week          mm.Value
		Low52week           mm.Value
		HighLow52WeekUpdate time.Time
	}

	// Price represents stock's price struct
	Price struct {
		Date   time.Time
		Close  float64
		High   float64
		Low    float64
		Open   float64
		Change float64
		Volume int64
	}

	// Price52WeekHighLow represents 52 week high -low stock's price struct
	Price52WeekHighLow struct {
		High float64
		Low  float64
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
