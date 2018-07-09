package service

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/dohernandez/go-quote"
	gf "github.com/dohernandez/googlefinance-client-go"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	StockPrice interface {
		Price(stk *stock.Stock) (stock.Price, error)
	}

	stockPrice struct {
		ctx       context.Context
		iexClient *iex.Client
	}
)

func NewStockPrice(ctx context.Context, iexClient *iex.Client) *stockPrice {
	return &stockPrice{
		ctx:       ctx,
		iexClient: iexClient,
	}
}

func (s *stockPrice) Price(stk *stock.Stock) (stock.Price, error) {
	method := "closedPriceFromYahoo"

	p, err := s.closedPriceFromYahoo(stk)
	if err != nil {
		logger.FromContext(s.ctx).WithError(err).Debugf("failed %s for stock %q", method, stk.Symbol)
		time.Sleep(5 * time.Second)
		method = "closedPriceFromGoogle"

		p, err = s.closedPriceFromGoogle(stk)
		if err != nil {
			logger.FromContext(s.ctx).WithError(err).Debugf("failed %s for stock %q", method, stk.Symbol)
			time.Sleep(5 * time.Second)
			method = "closedPriceFromIEXTrading"

			p, err = s.closedPriceFromIEXTrading(stk)
			if err != nil {
				if err == mm.ErrNotFound {
					return stock.Price{}, err
				}

				return stock.Price{}, errors.WithStack(err)
			}
		}
	}
	logger.FromContext(s.ctx).Debugf("got price %+v from stock %s with method %s", p, stk.Symbol, method)

	return p, nil
}

func (s *stockPrice) closedPriceFromYahoo(stk *stock.Stock) (stock.Price, error) {
	endDate := time.Now()
	startDate := endDate.Add(-24 * time.Hour)

	q, err := quote.NewQuoteFromYahoo(stk.Symbol, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), quote.Daily, true)
	if err != nil {
		return stock.Price{}, err
	}

	return stock.Price{
		Date:   q.Date[0],
		Close:  q.Close[0],
		High:   q.High[0],
		Low:    q.Low[0],
		Open:   q.Open[0],
		Volume: int64(q.Volume[0]),
		Change: q.Close[0] - q.Open[0],
	}, nil
}

func (s *stockPrice) closedPriceFromGoogle(stk *stock.Stock) (stock.Price, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gps, err := gf.GetPrices(ctx, &gf.Query{
		P: "2d",
		I: "86400",
		X: stk.Exchange.Symbol,
		Q: stk.Symbol,
	})
	if err != nil {
		return stock.Price{}, err
	}

	if len(gps) == 0 {
		return stock.Price{}, errors.Errorf("symbol '%s' not found\n", stk.Symbol)
	}

	p := gps[len(gps)-1]

	return stock.Price{
		Date:   p.Date,
		Close:  p.Close,
		High:   p.High,
		Low:    p.Low,
		Open:   p.Open,
		Volume: p.Volume,
		Change: p.Close - p.Open,
	}, nil
}

func (s *stockPrice) closedPriceFromIEXTrading(stk *stock.Stock) (stock.Price, error) {
	q, err := s.iexClient.Quote.Get(stk.Symbol)
	if err != nil {
		return stock.Price{}, err
	}
	return stock.Price{
		//Date:   q.LatestUpdate,
		Close:  q.Close,
		High:   q.High,
		Low:    q.Low,
		Open:   q.Open,
		Volume: q.Volume,
		Change: q.Close - q.Open,
	}, nil
}
