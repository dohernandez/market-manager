package service

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/dohernandez/go-quote"
	gf "github.com/dohernandez/googlefinance-client-go"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

const UpdatePriceConcurrency = 10

type (
	Purchase struct {
		ctx context.Context

		stockFinder         stock.Finder
		stockDividendFinder dividend.Finder
		marketFinder        market.Finder
		exchangeFinder      exchange.Finder
		stockInfoFinder     stock.InfoFinder

		stockPersister         stock.Persister
		stockDividendPersister dividend.Persister
		stockInfoPersister     stock.InfoPersister

		iexClient *iex.Client

		accountService *Account
	}
)

func NewPurchaseService(
	ctx context.Context,
	stockPersister stock.Persister,
	stockFinder stock.Finder,
	stockDividendPersister dividend.Persister,
	stockDividendFinder dividend.Finder,
	marketFinder market.Finder,
	exchangeFinder exchange.Finder,
	accountService *Account,
	iexClient *iex.Client,
	stockInfoFinder stock.InfoFinder,
	stockInfoPersister stock.InfoPersister,
) *Purchase {
	return &Purchase{
		ctx:                    ctx,
		stockPersister:         stockPersister,
		stockFinder:            stockFinder,
		stockDividendPersister: stockDividendPersister,
		stockDividendFinder:    stockDividendFinder,
		marketFinder:           marketFinder,
		exchangeFinder:         exchangeFinder,
		accountService:         accountService,
		iexClient:              iexClient,
		stockInfoFinder:        stockInfoFinder,
		stockInfoPersister:     stockInfoPersister,
	}
}

func (s *Purchase) FindMarketByName(name string) (*market.Market, error) {
	return s.marketFinder.FindByName(name)
}

func (s *Purchase) FindExchangeBySymbol(symbol string) (*exchange.Exchange, error) {
	return s.exchangeFinder.FindBySymbol(symbol)
}

func (s *Purchase) SaveAllStocks(stks []*stock.Stock) error {
	return s.stockPersister.PersistAll(stks)
}

func (s *Purchase) FindStockBySymbol(symbol string) (*stock.Stock, error) {
	return s.stockFinder.FindBySymbol(symbol)
}

func (s *Purchase) FindStockByName(name string) (*stock.Stock, error) {
	return s.stockFinder.FindByName(name)
}

func (s *Purchase) Stocks() ([]*stock.Stock, error) {
	stks, err := s.stockFinder.FindAll()
	if err != nil {
		return nil, err
	}

	for _, stk := range stks {
		d, err := s.stockDividendFinder.FindUpcoming(stk.ID)
		if err != nil {
			if err != mm.ErrNotFound {
				return nil, err
			}

			continue
		}

		stk.Dividends = append(stk.Dividends, d)
	}

	return stks, nil
}

func (s *Purchase) StocksByExchanges(exchanges []string) ([]*stock.Stock, error) {
	return s.stockFinder.FindAllByExchanges(exchanges)
}

// UpdateLastClosedPriceStocks update the stocks with the last close date price
func (s *Purchase) UpdateLastClosedPriceStocks(stks []*stock.Stock) []error {
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

func (s *Purchase) getLastClosedPriceOfStock(stk *stock.Stock) (stock.Price, error) {
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

func (s *Purchase) getLastClosedPriceFromYahoo(stk *stock.Stock) (stock.Price, error) {
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

func (s *Purchase) getLastClosedPriceFromGoogle(stk *stock.Stock) (stock.Price, error) {
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

func (s *Purchase) getLastClosedPriceFromIEXTrading(stk *stock.Stock) (stock.Price, error) {
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

func (s *Purchase) updateLastClosedPriceOfStock(stk *stock.Stock, p stock.Price) error {
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

	if stk.High52Week.Amount < p.High {
		stk.High52Week = mm.Value{
			Amount:   p.High,
			Currency: stk.High52Week.Currency,
		}
	}

	if stk.Low52Week.Amount > p.Low {
		stk.Low52Week = mm.Value{
			Amount:   p.High,
			Currency: stk.Low52Week.Currency,
		}
	}

	return s.stockPersister.UpdatePrice(stk)
}

func (s *Purchase) UpdateLastClosedPriceStock(stk *stock.Stock) error {
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
func (s *Purchase) Update52WeekHighLowPriceStocks(stks []*stock.Stock) []error {
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

func (s *Purchase) get52WeekHighLowPriceOfStock(stk *stock.Stock) (stock.Price52WeekHighLow, error) {
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

func (s *Purchase) get52WeekHighLowPriceFromYahoo(stk *stock.Stock) (stock.Price52WeekHighLow, error) {
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
		High52Week: high52wk,
		Low52Week:  low52wk,
	}, nil
}

func (s *Purchase) get52WeekHighLowPriceFromGoogle(stk *stock.Stock) (stock.Price52WeekHighLow, error) {
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
		High52Week: high52wk,
		Low52Week:  low52wk,
	}, nil

	p := gps[len(gps)-1]

	return stock.Price52WeekHighLow{
		High52Week: p.High,
		Low52Week:  p.Low,
	}, nil
}

func (s *Purchase) update52WeekHighLowPriceOfStock(stk *stock.Stock, p stock.Price52WeekHighLow) error {
	c := mm.ExchangeCurrency(stk.Exchange.Symbol)

	stk.High52Week = mm.Value{
		Amount:   p.High52Week,
		Currency: c,
	}

	stk.Low52Week = mm.Value{
		Amount:   p.Low52Week,
		Currency: c,
	}

	return s.stockPersister.UpdateHighLow52WeekPrice(stk)
}

func (s *Purchase) Update52WeekHighLowPriceStock(stk *stock.Stock) error {
	p, err := s.get52WeekHighLowPriceOfStock(stk)
	if err != nil {
		return err
	}

	if err := s.update52WeekHighLowPriceOfStock(stk, p); err != nil {
		return err
	}

	return nil
}

func (s *Purchase) UpdateStockDividends(stk *stock.Stock) error {
	err := s.stockDividendPersister.PersistAll(stk.ID, stk.Dividends)
	if err != nil {
		return err
	}

	return s.updateStockDividendYield(stk)
}

func (s *Purchase) updateStockDividendYield(stk *stock.Stock) error {
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

func (s *Purchase) StocksByDividendAnnounceProjectYearAndMonth(year, month string) ([]*stock.Stock, error) {
	y, _ := strconv.Atoi(year)
	m, _ := strconv.Atoi(month)

	return s.stockFinder.FindAllByDividendAnnounceProjectYearAndMonth(y, m)
}

func (s *Purchase) FindStockInfoByValue(symbol string) (*stock.Info, error) {
	return s.stockInfoFinder.FindByName(symbol)
}

func (s *Purchase) SaveStockInfo(stkInfo *stock.Info) error {
	return s.stockInfoPersister.Persist(stkInfo)
}
