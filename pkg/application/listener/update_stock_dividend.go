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

const (
	updateDividendConcurrency = 5

	updateDividendSleep time.Duration = 15
)

type updateStockDividend struct {
	stockDividendPersister dividend.Persister
	stockDividendService   service.StockDividend
	startHistoricalDate    time.Time

	concurrency int
	sleep       time.Duration
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
		concurrency:            updateDividendConcurrency,
		sleep:                  updateDividendSleep,
	}
}

func (l *updateStockDividend) OnEvent(ctx context.Context, event cbus.Event) {
	stks := event.Result.([]*stock.Stock)

	if l.concurrency > 1 {
		l.updateDividendConcurrency(ctx, stks)
	} else {
		for _, stk := range stks {
			l.updateDividend(ctx, stk)
		}
	}

	logger.FromContext(ctx).Debug("Updated stock dividend")
}

func (l *updateStockDividend) updateDividendConcurrency(
	ctx context.Context,
	stks []*stock.Stock,
) {
	var wg sync.WaitGroup

	concurrency := l.concurrency

	for _, stk := range stks {
		wg.Add(1)

		concurrency--
		st := stk

		go func() {
			defer wg.Done()

			l.updateDividend(ctx, st)

			concurrency++
		}()

		for {
			if concurrency != 0 {
				break
			}

			logger.FromContext(ctx).Errorf("Going to rest for %d seconds", 15)
			time.Sleep(l.sleep * time.Second)
			logger.FromContext(ctx).Errorf("Waking up after %d seconds sleeping", 15)
		}
	}

	wg.Wait()
}

func (l *updateStockDividend) updateDividend(ctx context.Context, stk *stock.Stock) {
	err := l.stockDividendPersister.DeleteAll(stk.ID)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while deleting all stock dividend symbol [%s] -> error [%s]",
			stk.Symbol,
			err,
		)

		return
	}

	var ds []dividend.StockDividend

	dsf, err := l.stockDividendService.Future(stk)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while updating all stock dividend future symbol [%s] -> error [%s]",
			stk.Symbol,
			err,
		)
	} else {
		ds = append(ds, dsf...)
	}

	dsh, err := l.stockDividendService.Historical(stk, l.startHistoricalDate)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while updating all stock dividend historical symbol [%s] -> error [%s]",
			stk.Symbol,
			err,
		)
	} else {
		ds = append(ds, dsh...)
	}

	if len(ds) > 0 {
		stk.Dividends = ds
		err = l.stockDividendPersister.PersistAll(stk.ID, ds)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while updating all stock dividend symbol [%s] -> error [%s]",
				stk.Symbol,
				err,
			)
		}
	}
}

func (l *updateStockDividend) WithConcurrency(concurrency int) {
	l.concurrency = concurrency
}

func (l *updateStockDividend) WithSleep(sleep time.Duration) {
	l.sleep = sleep
}
