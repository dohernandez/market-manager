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
	InfoType string

	Info struct {
		ID   uuid.UUID
		Type InfoType
		Name string
	}

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
		High52Week          mm.Value
		Low52Week           mm.Value
		HighLow52WeekUpdate time.Time
		Type                *Info
		Sector              *Info
		Industry            *Info
	}

	// Price represents stock's price struct
	Price struct {
		Date       time.Time
		Close      float64
		High       float64
		Low        float64
		Open       float64
		Change     float64
		Volume     int64
		High52Week float64
		Low52Week  float64
	}

	// Price52WeekHighLow represents 52 week high -low stock's price struct
	Price52WeekHighLow struct {
		High52Week float64
		Low52Week  float64
	}
)

const (
	StockInfoType     InfoType = "type"
	StockInfoSector            = "sector"
	StockInfoIndustry          = "industry"
)

// NewStockInfo creates an stock info instance
func NewStockInfo(name string, t InfoType) *Info {
	return &Info{
		ID:   uuid.NewV4(),
		Name: name,
		Type: t,
	}
}

// NewStock creates an stock instance
func NewStock(market *market.Market, exchange *exchange.Exchange, name, symbol string, t, sector, industry *Info) *Stock {
	return &Stock{
		ID:                  uuid.NewV4(),
		Market:              market,
		Exchange:            exchange,
		Name:                name,
		Symbol:              symbol,
		Type:                t,
		Sector:              sector,
		Industry:            industry,
		LastPriceUpdate:     time.Time{},
		HighLow52WeekUpdate: time.Time{},
	}
}

// ComparePriceWithHighLow Compare price with stock high - low price and returns
// 1 - Price between 71% - 100%
// 0 - Price between 31% - 70%
// -1 - Price between 0% - 30%
func (s *Stock) ComparePriceWithHighLow() int {
	r52wk := s.High52Week.Decrease(s.Low52Week)
	p := s.Value

	if p.Amount > s.High52Week.Amount-r52wk.Amount/3 {
		return 1
	}

	if p.Amount > s.Low52Week.Amount+r52wk.Amount/3 {
		return 0
	}

	return -1
}

// BuyUnder Price proposal when is appropriate to buy the stock
func (s *Stock) BuyUnder() mm.Value {
	r52wk := s.High52Week.Decrease(s.Low52Week)

	return mm.Value{
		Amount:   s.High52Week.Amount - r52wk.Amount/3*2,
		Currency: r52wk.Currency,
	}
}

func (s *Stock) Equals(stk *Stock) bool {
	return s.Symbol == stk.Symbol
}
