package handler

import (
	"context"

	"github.com/gogolfing/cbus"

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
	return nil, nil
}
