package stock

import (
	"context"
	"sync"
	"time"

	"fmt"

	quote "github.com/dohernandez/go-quote"
	gf "github.com/dohernandez/googlefinance-client-go"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock/dividend"
)

type (
	Service struct {
		stockPersister         Persister
		stockDividendPersister dividend.Persister
		stockFinder            Finder
		stockDividendFinder    dividend.Finder
	}
)

func NewService(stockPersister Persister, stockFinder Finder, stockDividendPersister dividend.Persister, stockDividendFinder dividend.Finder) *Service {
	return &Service{
		stockPersister:         stockPersister,
		stockFinder:            stockFinder,
		stockDividendPersister: stockDividendPersister,
		stockDividendFinder:    stockDividendFinder,
	}
}

func (s *Service) FindStockBySymbol(symbol string) (*Stock, error) {
	return s.stockFinder.FindBySymbol(symbol)
}

func (s *Service) Stocks() ([]*Stock, error) {
	return s.stockFinder.FindAll()
}

func (s *Service) SaveAll(stks []*Stock) error {
	return s.stockPersister.PersistAll(stks)
}

func (s *Service) GetLastClosedPriceFromGoogle(stk *Stock) (Price, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gps, err := gf.GetPrices(ctx, &gf.Query{
		P: "2d",
		I: "86400",
		X: stk.Exchange.Symbol,
		Q: stk.Symbol,
	})
	if err != nil {
		return Price{}, err
	}

	if len(gps) == 0 {
		return Price{}, errors.New(fmt.Sprintf("symbol '%s' not found\n", stk.Symbol))
	}

	p := gps[len(gps)-1]

	return Price{
		Date:   p.Date,
		Close:  p.Close,
		High:   p.High,
		Low:    p.Low,
		Open:   p.Open,
		Volume: float64(p.Volume),
	}, nil
}

func (s *Service) GetLastClosedPriceFromYahoo(stk *Stock) (Price, error) {
	endDate := time.Now()
	startDate := endDate.Add(-24 * time.Hour)

	q, err := quote.NewQuoteFromYahoo(stk.Symbol, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), quote.Daily, true)
	if err != nil {
		return Price{}, err
	}

	return Price{
		Date:   q.Date[0],
		Close:  q.Close[0],
		High:   q.High[0],
		Low:    q.Low[0],
		Open:   q.Open[0],
		Volume: q.Volume[0],
	}, nil
}

func (s *Service) UpdateLastClosedPriceStocks(stks []*Stock) []error {
	var (
		wg   sync.WaitGroup
		errs []error
	)

	for _, stk := range stks {
		wg.Add(1)

		st := stk
		go func() {
			defer wg.Done()

			if err := s.updateLastClosedPriceOfStock(st); err != nil {
				errs = append(errs, errors.New(fmt.Sprintf("%+v -> stock:%+v", err, st)))
			}
		}()
	}

	wg.Wait()

	return errs
}

func (s *Service) updateLastClosedPriceOfStock(stk *Stock) error {
	p, err := s.GetLastClosedPriceFromYahoo(stk)
	if err != nil {
		p, err = s.GetLastClosedPriceFromGoogle(stk)
		if err != nil {
			return err
		}
	}

	stk.Value = mm.Value{
		Amount:   p.Close,
		Currency: mm.Dollar,
	}

	err = s.updateStockDividendYield(stk)
	if err != nil {
		return err
	}

	return s.stockPersister.UpdatePrice(stk)
}

func (s *Service) UpdateLastClosedPriceStock(stk *Stock) error {
	if err := s.updateLastClosedPriceOfStock(stk); err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateStockDividends(stk *Stock) error {
	err := s.stockDividendPersister.PersistAll(stk.ID, stk.Dividends)
	if err != nil {
		return err
	}

	return s.updateStockDividendYield(stk)
}

func (s *Service) updateStockDividendYield(stk *Stock) error {
	d, err := s.stockDividendFinder.FindNextFromStock(stk.ID, time.Now())
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		return nil
	}

	if stk.Value.Amount <= 0 {
		return errors.New("stock value is 0 or less that 0")
	}

	stk.DividendYield = d.Amount * 4 / stk.Value.Amount * 100

	return s.stockPersister.UpdateDividendYield(stk)
}
