package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type updateAllStockDividend struct {
	stockFinder stock.Finder
}

func NewUpdateAllStockDividend(stockFinder stock.Finder) *updateAllStockDividend {
	return &updateAllStockDividend{
		stockFinder: stockFinder,
	}
}

func (h *updateAllStockDividend) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	stks, err := h.stockFinder.FindAll()
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while finding all stock -> error [%s]",
			err,
		)

		return nil, err
	}

	return stks, nil
}
