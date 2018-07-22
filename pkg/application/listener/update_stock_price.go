package listener

import (
	"context"
	"sync"
	"time"

	"github.com/gogolfing/cbus"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

const updatePriceConcurrency = 10

type updateStockPrice struct {
	stockFinder       stock.Finder
	stockPriceService service.StockPrice
	stockPersister    stock.Persister
}

func NewUpdateStockPrice(stockFinder stock.Finder, stockPriceService service.StockPrice, stockPersister stock.Persister) *updateStockPrice {
	return &updateStockPrice{
		stockFinder:       stockFinder,
		stockPriceService: stockPriceService,
		stockPersister:    stockPersister,
	}
}

func (l *updateStockPrice) OnEvent(ctx context.Context, event cbus.Event) {
	stks := event.Result.([]*stock.Stock)

	var (
		wg   sync.WaitGroup
		ustk []*stock.Stock
	)

	concurrency := updatePriceConcurrency
	for _, stk := range stks {
		wg.Add(1)
		concurrency--

		st := stk
		go func() {
			defer wg.Done()

			err := l.updateStock(st)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while updating stocks price: stock [%s] -> error [%s]",
					st.Symbol,
					err,
				)

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

	logger.FromContext(ctx).Debug("Updated stocks price")
}

func (l *updateStockPrice) updateStock(stk *stock.Stock) error {
	p, err := l.stockPriceService.Price(stk)
	if err != nil {
		return errors.Wrapf(err, "symbol : %s", stk.Symbol)
	}

	stk.Value = mm.Value{
		Amount:   p.Close,
		Currency: mm.Dollar,
	}

	stk.Change = mm.Value{
		Amount:   p.Change,
		Currency: mm.Dollar,
	}

	if p.High52Week > 0 {
		stk.High52Week = mm.Value{
			Amount:   p.High52Week,
			Currency: stk.High52Week.Currency,
		}
	} else if stk.High52Week.Amount < p.High {
		stk.High52Week = mm.Value{
			Amount:   p.High,
			Currency: stk.High52Week.Currency,
		}
	}

	if p.Low52Week > 0 {
		stk.Low52Week = mm.Value{
			Amount:   p.Low52Week,
			Currency: stk.Low52Week.Currency,
		}
	} else if stk.Low52Week.Amount > p.Low {
		stk.Low52Week = mm.Value{
			Amount:   p.High,
			Currency: stk.Low52Week.Currency,
		}
	}

	stk.EPS = p.EPS
	stk.PER = p.PER

	if err := l.stockPersister.UpdatePrice(stk); err != nil {
		return errors.Wrapf(err, "symbol : %s", stk.Symbol)
	}

	return nil
}
