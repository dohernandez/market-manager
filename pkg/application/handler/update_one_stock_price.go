package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type updateOneStockPrice struct {
	stockFinder stock.Finder
}

func NewUpdateOneStockPrice(stockFinder stock.Finder) *updateOneStockPrice {
	return &updateOneStockPrice{
		stockFinder: stockFinder,
	}
}

func (h *updateOneStockPrice) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	symbol := command.(*appCommand.UpdateOneStockPrice).Symbol

	stk, err := h.stockFinder.FindBySymbol(symbol)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while finding stock: symbol [%s] -> error [%s]",
			symbol,
			err,
		)

		return nil, err
	}

	return []*stock.Stock{
		stk,
	}, nil
}
