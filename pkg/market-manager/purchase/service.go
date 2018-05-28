package purchase

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	quote "github.com/dohernandez/go-quote"
	gf "github.com/dohernandez/googlefinance-client-go"

	"fmt"

	"github.com/dohernandez/market-manager/pkg/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	Service struct {
		stockFinder         stock.Finder
		stockDividendFinder dividend.Finder
		marketFinder        market.Finder
		exchangeFinder      exchange.Finder

		stockPersister         stock.Persister
		stockDividendPersister dividend.Persister

		iexClient *iex.Client

		accountService *account.Service
	}
)

func NewService(
	stockPersister stock.Persister,
	stockFinder stock.Finder,
	stockDividendPersister dividend.Persister,
	stockDividendFinder dividend.Finder,
	marketFinder market.Finder,
	exchangeFinder exchange.Finder,
	accountService *account.Service,
	iexClient *iex.Client,
) *Service {
	return &Service{
		stockPersister:         stockPersister,
		stockFinder:            stockFinder,
		stockDividendPersister: stockDividendPersister,
		stockDividendFinder:    stockDividendFinder,
		marketFinder:           marketFinder,
		exchangeFinder:         exchangeFinder,
		accountService:         accountService,
		iexClient:              iexClient,
	}
}

func (s *Service) FindMarketByName(name string) (*market.Market, error) {
	return s.marketFinder.FindByName(name)
}

func (s *Service) FindExchangeBySymbol(symbol string) (*exchange.Exchange, error) {
	return s.exchangeFinder.FindBySymbol(symbol)
}

func (s *Service) SaveAllStocks(stks []*stock.Stock) error {
	return s.stockPersister.PersistAll(stks)
}

func (s *Service) FindStockBySymbol(symbol string) (*stock.Stock, error) {
	return s.stockFinder.FindBySymbol(symbol)
}

func (s *Service) FindStockByName(name string) (*stock.Stock, error) {
	return s.stockFinder.FindByName(name)
}

func (s *Service) Stocks() ([]*stock.Stock, error) {
	return s.stockFinder.FindAll()
}

func (s *Service) UpdateLastClosedPriceStocks(stks []*stock.Stock) []error {
	var (
		wg    sync.WaitGroup
		errs  []error
		calls int
		ustk  []*stock.Stock
	)

	for _, stk := range stks {
		wg.Add(1)

		st := stk
		go func() {
			defer wg.Done()

			calls++

			p, err := s.getLastClosedPriceOfStock(st)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "symbol : %s", st.Symbol))

				return
			}

			if err := s.updateLastClosedPriceOfStock(st, p); err != nil {
				errs = append(errs, errors.Wrapf(err, "symbol : %s", st.Symbol))

				return
			}

			ustk = append(ustk, st)

		}()

		if calls >= 30 {
			fmt.Printf("Going to sleep, due to exede amount of call to the api")
			time.Sleep(time.Second * 60)

			calls = 0
		}
	}

	wg.Wait()

	err := s.accountService.UpdateWalletsCapitalByStocks(ustk)
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (s *Service) getLastClosedPriceOfStock(stk *stock.Stock) (stock.Price, error) {
	p, err := s.getLastClosedPriceFromYahoo(stk)
	if err != nil {
		p, err = s.getLastClosedPriceFromGoogle(stk)
		if err != nil {
			p, err = s.getLastClosedPriceFromIEXTrading(stk)
			if err != nil {
				if err == mm.ErrNotFound {
					return stock.Price{}, err
				}

				return stock.Price{}, errors.WithStack(err)
			}
		}
	}

	return p, nil
}

func (s *Service) getLastClosedPriceFromYahoo(stk *stock.Stock) (stock.Price, error) {
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
		Volume: q.Volume[0],
	}, nil
}

func (s *Service) getLastClosedPriceFromGoogle(stk *stock.Stock) (stock.Price, error) {
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
		Volume: float64(p.Volume),
	}, nil
}

func (s *Service) getLastClosedPriceFromIEXTrading(stk *stock.Stock) (stock.Price, error) {
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
	}, nil
}

func (s *Service) updateLastClosedPriceOfStock(stk *stock.Stock, p stock.Price) error {
	pv := stk.Value

	stk.Value = mm.Value{
		Amount:   p.Close,
		Currency: mm.Dollar,
	}

	stk.Change = stk.Value.Decrease(pv)

	err := s.updateStockDividendYield(stk)
	if err != nil {
		return err
	}

	return s.stockPersister.UpdatePrice(stk)
}

func (s *Service) UpdateLastClosedPriceStock(stk *stock.Stock) error {
	p, err := s.getLastClosedPriceOfStock(stk)
	if err != nil {
		return err
	}

	if err := s.updateLastClosedPriceOfStock(stk, p); err != nil {
		return err
	}

	err = s.accountService.UpdateWalletsCapitalByStock(stk)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateStockDividends(stk *stock.Stock) error {
	err := s.stockDividendPersister.PersistAll(stk.ID, stk.Dividends)
	if err != nil {
		return err
	}

	return s.updateStockDividendYield(stk)
}

func (s *Service) updateStockDividendYield(stk *stock.Stock) error {
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
