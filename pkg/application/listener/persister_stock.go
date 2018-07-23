package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type persisterStock struct {
	stockPersister stock.Persister
}

func NewPersisterStock(stockPersister stock.Persister) *persisterStock {
	return &persisterStock{
		stockPersister: stockPersister,
	}
}

func (l *persisterStock) OnEvent(ctx context.Context, event cbus.Event) {
	stks := event.Result.([]*stock.Stock)

	err := l.stockPersister.PersistAll(stks)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting stocks -> error [%s]",
			err,
		)
	}
}
