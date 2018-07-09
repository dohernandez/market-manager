package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	"sync"
	"time"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

const UpdatePriceConcurrency = 10

type updateAllStockPrice struct {
	updateStockPrice
}

func NewUpdateAllStockPrice(stockFinder stock.Finder, stockPriceService service.StockPrice, stockPersister stock.Persister) *updateAllStockPrice {
	return &updateAllStockPrice{
		updateStockPrice: updateStockPrice{
			stockFinder:       stockFinder,
			stockPriceService: stockPriceService,
			stockPersister:    stockPersister,
		},
	}
}

func (h *updateAllStockPrice) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	stks, err := h.stockFinder.FindAll()
	if err != nil {
		return nil, err
	}

	var (
		wg   sync.WaitGroup
		ustk []*stock.Stock
	)

	concurrency := UpdatePriceConcurrency
	for _, stk := range stks {
		wg.Add(1)
		concurrency--

		st := stk
		go func() {
			defer wg.Done()

			err := h.updateStock(st)
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

	return ustk, nil
}
