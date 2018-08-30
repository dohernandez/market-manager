package listener

import (
	"context"
	"sync"
	"time"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

const updateDividendConcurrency = 5

type updateStockDividend struct {
	stockDividendPersister dividend.Persister
	stockDividendService   service.StockDividend
	startHistoricalDate    time.Time
}

func NewUpdateStockDividend(
	stockDividendPersister dividend.Persister,
	stockDividendService service.StockDividend,
) *updateStockDividend {
	startHistoricalDate, _ := time.Parse("2-Jan-2006", "1-Jan-2017")

	return &updateStockDividend{
		stockDividendPersister: stockDividendPersister,
		stockDividendService:   stockDividendService,
		startHistoricalDate:    startHistoricalDate,
	}
}

func (l *updateStockDividend) OnEvent(ctx context.Context, event cbus.Event) {
	stks := event.Result.([]*stock.Stock)

	var wg sync.WaitGroup

	concurrency := updateDividendConcurrency

	for _, stk := range stks {
		wg.Add(1)
		concurrency--

		st := stk
		go func() {
			defer wg.Done()

			err := l.stockDividendPersister.DeleteAll(st.ID)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while deleting all stock dividend symbol [%s] -> error [%s]",
					stk.Symbol,
					err,
				)

				concurrency++

				return
			}

			var ds []dividend.StockDividend

			dsf, err := l.stockDividendService.Future(st)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while updating all stock dividend future symbol [%s] -> error [%s]",
					st.Symbol,
					err,
				)
			} else {
				ds = append(ds, dsf...)
			}

			dsh, err := l.stockDividendService.Historical(st, l.startHistoricalDate)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while updating all stock dividend historical symbol [%s] -> error [%s]",
					st.Symbol,
					err,
				)
			} else {
				ds = append(ds, dsh...)
			}

			st.Dividends = ds
			err = l.stockDividendPersister.PersistAll(st.ID, ds)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while updating all stock dividend symbol [%s] -> error [%s]",
					st.Symbol,
					err,
				)
			}

			concurrency++
		}()

		for {
			if concurrency != 0 {
				break
			}

			logger.FromContext(ctx).Errorf("Going to rest for %d seconds", 15)
			time.Sleep(15 * time.Second)
			logger.FromContext(ctx).Errorf("Waking up after %d seconds sleeping", 15)
		}
	}

	wg.Wait()

	logger.FromContext(ctx).Debug("Updated stock dividend")
}
