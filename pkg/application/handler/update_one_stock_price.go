package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type updateOneStockPrice struct {
	updateStockPrice
}

func NewUpdateOneStockPrice(stockFinder stock.Finder, stockPriceService service.StockPrice, stockPersister stock.Persister) *updateOneStockPrice {
	return &updateOneStockPrice{
		updateStockPrice: updateStockPrice{
			stockFinder:       stockFinder,
			stockPriceService: stockPriceService,
			stockPersister:    stockPersister,
		},
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

	var ustk []*stock.Stock

	err = h.updateStock(stk)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while updating stocks price: stock [%s] -> error [%s]",
			stk.Symbol,
			err,
		)

		return nil, err
	}

	ustk = append(ustk, stk)

	return ustk, nil
}
