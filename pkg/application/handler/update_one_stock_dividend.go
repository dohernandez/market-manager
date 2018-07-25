package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type updateOneStockDividend struct {
	stockFinder stock.Finder
}

func NewUpdateOneStockDividend(stockFinder stock.Finder) *updateOneStockDividend {
	return &updateOneStockDividend{
		stockFinder: stockFinder,
	}
}

func (h *updateOneStockDividend) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	symbol := command.(*appCommand.UpdateOneStockDividend).Symbol

	var stks []*stock.Stock

	stk, err := h.stockFinder.FindBySymbol(symbol)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while finding stock symbol [%s]-> error [%s]",
			symbol,
			err,
		)

		return nil, err
	}

	stks = append(stks, stk)

	return stks, nil
}
