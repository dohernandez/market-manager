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

// ComparePriceWithHighLow Compare price with stock high - low price and returns
// 1 - Price between 71% - 100%
// 0 - Price between 31% - 70%
// -1 - Price between 0% - 30%
func (s *Stock) ComparePriceWithHighLow() int {
	r52wk := s.High52week.Decrease(s.Low52week)
	p := s.Value

	if p.Amount > s.High52week.Amount-r52wk.Amount/3 {
		return 1
	}

	if p.Amount > s.Low52week.Amount+r52wk.Amount/3 {
		return 0
	}

	return -1
}

// BuyUnder Price proposal when is appropriate to buy the stock
func (s *Stock) BuyUnder() mm.Value {
	r52wk := s.High52week.Decrease(s.Low52week)

	return mm.Value{
		Amount:   s.High52week.Amount - r52wk.Amount/3*2,
		Currency: r52wk.Currency,
	}
}
