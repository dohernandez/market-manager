package purchase

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/dohernandez/go-quote"
	gf "github.com/dohernandez/googlefinance-client-go"

	"strconv"

	"github.com/dohernandez/market-manager/pkg/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

const UpdatePriceConcurrency = 10

type (
	Service struct {
		ctx context.Context

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
	ctx context.Context,
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
		ctx:                    ctx,
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

func (s *Service) StocksByExchanges(exchanges []string) ([]*stock.Stock, error) {
	return s.stockFinder.FindAllByExchanges(exchanges)
}

// UpdateLastClosedPriceStocks update the stocks with the last close date price
func (s *Service) UpdateLastClosedPriceStocks(stks []*stock.Stock) []error {
	var (
		wg   sync.WaitGroup
		errs []error
		ustk []*stock.Stock
	)

	concurrency := UpdatePriceConcurrency
	for _, stk := range stks {
		wg.Add(1)
		concurrency--

		st := stk
		go func() {
			defer wg.Done()

			p, err := s.getLastClosedPriceOfStock(st)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "symbol : %s", st.Symbol))

				concurrency++
				return
			}

			if err := s.updateLastClosedPriceOfStock(st, p); err != nil {
				errs = append(errs, errors.Wrapf(err, "symbol : %s", st.Symbol))

				concurrency++
				return
			}

			ustk = append(ustk, st)
			concurrency++
		}()

		for {
			if concurrency != 0 {
				break
			}
			time.Sleep(2 * time.Second)
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
	method := "getLastClosedPriceFromYahoo"

	p, err := s.getLastClosedPriceFromYahoo(stk)
	if err != nil {
		logger.FromContext(s.ctx).WithError(err).Debugf("failed %s for stock %q", method, stk.Symbol)
		time.Sleep(5 * time.Second)
		method = "getLastClosedPriceFromGoogle"

		p, err = s.getLastClosedPriceFromGoogle(stk)
		if err != nil {
			logger.FromContext(s.ctx).WithError(err).Debugf("failed %s for stock %q", method, stk.Symbol)
			time.Sleep(5 * time.Second)
			method = "getLastClosedPriceFromIEXTrading"

			p, err = s.getLastClosedPriceFromIEXTrading(stk)
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
		Volume: int64(q.Volume[0]),
		Change: q.Close[0] - q.Open[0],
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
		Volume: p.Volume,
		Change: p.Close - p.Open,
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
		Change: q.Close - q.Open,
	}, nil
}

func (s *Service) updateLastClosedPriceOfStock(stk *stock.Stock, p stock.Price) error {
	stk.Value = mm.Value{
		Amount:   p.Close,
		Currency: mm.Dollar,
	}

	stk.Change = mm.Value{
		Amount:   p.Change,
		Currency: mm.Dollar,
	}

	err := s.updateStockDividendYield(stk)
	if err != nil {
		return err
	}

	if stk.High52week.Amount < p.High {
		stk.High52week = mm.Value{
			Amount:   p.High,
			Currency: stk.High52week.Currency,
		}
	}

	if stk.Low52week.Amount > p.Low {
		stk.Low52week = mm.Value{
			Amount:   p.High,
			Currency: stk.Low52week.Currency,
		}
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

// Update52WeekClosedPriceStocks update the stocks 52 week with the high - low price
func (s *Service) Update52WeekHighLowPriceStocks(stks []*stock.Stock) []error {
	var (
		wg   sync.WaitGroup
		errs []error
	)

	concurrency := UpdatePriceConcurrency
	for _, stk := range stks {
		wg.Add(1)
		concurrency--

		st := stk
		go func() {
			defer wg.Done()

			p, err := s.get52WeekHighLowPriceOfStock(st)
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "symbol : %s", st.Symbol))

				concurrency++
				return
			}

			if err := s.update52WeekHighLowPriceOfStock(st, p); err != nil {
				errs = append(errs, errors.Wrapf(err, "symbol : %s", st.Symbol))

				concurrency++
				return
			}

			concurrency++
		}()

		for {
			if concurrency != 0 {
				break
			}
			time.Sleep(2 * time.Second)
		}
	}

	wg.Wait()

	return errs
}

func (s *Service) get52WeekHighLowPriceOfStock(stk *stock.Stock) (stock.Price52WeekHighLow, error) {
	method := "get52WeekHighLowPriceFromYahoo"

	p, err := s.get52WeekHighLowPriceFromYahoo(stk)
	if err != nil {
		logger.FromContext(s.ctx).WithError(err).Debugf("failed %s for stock %q", method, stk.Symbol)
		time.Sleep(5 * time.Second)
		method = "get52WeekHighLowPriceFromGoogle"

		p, err = s.get52WeekHighLowPriceFromGoogle(stk)
		if err != nil {
			if err == mm.ErrNotFound {
				return stock.Price52WeekHighLow{}, err
			}

			return stock.Price52WeekHighLow{}, errors.WithStack(err)
		}
	}
	logger.FromContext(s.ctx).Debugf("got 52 wk %+v from stock %s with method %s", p, stk.Symbol, method)

	return p, nil
}

func (s *Service) get52WeekHighLowPriceFromYahoo(stk *stock.Stock) (stock.Price52WeekHighLow, error) {
	endDate := time.Now()
	// startDate 52 week backward
	startDate := endDate.Add(-52 * 7 * 24 * time.Hour)

	q, err := quote.NewQuoteFromYahoo(stk.Symbol, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), quote.Daily, true)
	if err != nil {
		return stock.Price52WeekHighLow{}, err
	}

	high52wk := q.High[0]
	low52wk := q.Low[0]
	for k := range q.Date[1:] {
		if high52wk < q.High[k] {
			high52wk = q.High[k]
		}

		if low52wk > q.Low[k] {
			low52wk = q.Low[k]
		}
	}

	return stock.Price52WeekHighLow{
		High: high52wk,
		Low:  low52wk,
	}, nil
}

func (s *Service) get52WeekHighLowPriceFromGoogle(stk *stock.Stock) (stock.Price52WeekHighLow, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gps, err := gf.GetPrices(ctx, &gf.Query{
		P: "1Y",
		I: "86400",
		X: stk.Exchange.Symbol,
		Q: stk.Symbol,
	})
	if err != nil {
		return stock.Price52WeekHighLow{}, err
	}

	if len(gps) == 0 {
		return stock.Price52WeekHighLow{}, errors.Errorf("symbol '%s' not found\n", stk.Symbol)
	}

	high52wk := gps[0].High
	low52wk := gps[0].Low
	for _, gp := range gps[1:] {
		if high52wk < gp.High {
			high52wk = gp.High
		}

		if low52wk > gp.Low {
			low52wk = gp.Low
		}
	}

	return stock.Price52WeekHighLow{
		High: high52wk,
		Low:  low52wk,
	}, nil

	p := gps[len(gps)-1]

	return stock.Price52WeekHighLow{
		High: p.High,
		Low:  p.Low,
	}, nil
}

func (s *Service) update52WeekHighLowPriceOfStock(stk *stock.Stock, p stock.Price52WeekHighLow) error {
	c := mm.ExchangeCurrency(stk.Exchange.Symbol)

	stk.High52week = mm.Value{
		Amount:   p.High,
		Currency: c,
	}

	stk.Low52week = mm.Value{
		Amount:   p.Low,
		Currency: c,
	}

	return s.stockPersister.UpdateHighLow52WeekPrice(stk)
}

func (s *Service) Update52WeekHighLowPriceStock(stk *stock.Stock) error {
	p, err := s.get52WeekHighLowPriceOfStock(stk)
	if err != nil {
		return err
	}

	if err := s.update52WeekHighLowPriceOfStock(stk, p); err != nil {
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

	stk.DividendYield = d.Amount.Amount * 4 / stk.Value.Amount * 100

	return s.stockPersister.UpdateDividendYield(stk)
}

func (s *Service) StocksByDividendAnnounceProjectYearAndMonth(year, month string) ([]*stock.Stock, error) {
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)

	return s.stockFinder.FindAllByDividendAnnounceProjectYearAndMonth(y, m)
}
