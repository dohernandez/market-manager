package listener

import (
	"context"
	"sync"
	"time"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

const updatePriceVolatilityConcurrency = 15

type updateStockPriceVolatility struct {
	stockPriceVolatilityService service.StockPriceVolatility
	stockPersister              stock.Persister
}

func NewUpdateStockPriceVolatility(stockPriceVolatilityService service.StockPriceVolatility, stockPersister stock.Persister) *updateStockPriceVolatility {
	return &updateStockPriceVolatility{
		stockPriceVolatilityService: stockPriceVolatilityService,
		stockPersister:              stockPersister,
	}
}

func (l *updateStockPriceVolatility) OnEvent(ctx context.Context, event cbus.Event) {
	stks := event.Result.([]*stock.Stock)
	var (
		wg sync.WaitGroup
	)

	concurrency := updatePriceVolatilityConcurrency
	for _, stk := range stks {
		wg.Add(1)
		concurrency--

		st := stk
		go func() {
			defer wg.Done()

			pv, err := l.stockPriceVolatilityService.PriceVolatility(st)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while updating stocks price volatility: stock [%s] -\\u003e error [%s]",
					stk.Symbol,
					err,
				)

				concurrency++

				return
			}

			st.HV20Day = pv.HV20Day
			st.HV52Week = pv.HV52Week

			l.stockPersister.UpdatePriceVolatility(st)

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

	logger.FromContext(ctx).Debug("Updated stocks price volatility")
}
