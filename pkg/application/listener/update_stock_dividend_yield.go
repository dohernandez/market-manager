package listener

import (
	"context"

	"time"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type updateStockDividendYield struct {
	stockDividendFinder dividend.Finder
	stockPersister      stock.Persister
}

func NewUpdateStockDividendYield(stockDividendFinder dividend.Finder, stockPersister stock.Persister) *updateStockDividendYield {
	return &updateStockDividendYield{
		stockDividendFinder: stockDividendFinder,
		stockPersister:      stockPersister,
	}
}

func (l *updateStockDividendYield) OnEvent(ctx context.Context, event cbus.Event) {
	stks := event.Result.([]*stock.Stock)

	for _, stk := range stks {
		d, err := l.stockDividendFinder.FindNextFromStock(stk.ID, time.Now())

		if err != nil {
			if err != mm.ErrNotFound {
				logger.FromContext(ctx).Errorf(
					"An error happen while updating stocks dividend yield: stock [%s] -> error [%s]",
					stk.Symbol,
					err,
				)

				continue
			}

			continue
		}

		if stk.Value.Amount <= 0 {
			logger.FromContext(ctx).Errorf(
				"An error happen while updating stocks dividend yield: stock [%s] -> value is 0 or less that 0",
				stk.Symbol,
			)

			continue
		}

		stk.DividendYield = d.Amount.Amount * 4 / stk.Value.Amount * 100

		l.stockPersister.UpdateDividendYield(stk)
	}

	logger.FromContext(ctx).Debug("Updated stocks dividend yield")
}
