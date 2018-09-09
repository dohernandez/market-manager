package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type addStock struct {
	marketFinder   market.Finder
	exchangeFinder exchange.Finder
}

func NewAddStock() *addStock {
	return &addStock{}
}

func (h *addStock) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	exchangeSymbol := command.(*appCommand.AddStock).Exchange

	m, err := h.marketFinder.FindByName(market.Stock)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading market %s - error [%s]",
			market.Stock,
			err,
		)

		return nil, err
	}

	e, err := h.exchangeFinder.FindBySymbol(exchangeSymbol)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading exchange %s - error [%s]",
			exchangeSymbol,
			err,
		)

		return nil, err
	}

	return []*stock.Stock{
		stock.NewStockFromSymbol(m, e, command.(*appCommand.AddStock).Symbol),
	}, nil
}
