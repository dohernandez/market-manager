package handler

import (
	"context"
	"strings"
	"time"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/render"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type listStocks struct {
	stockFinder         stock.Finder
	stockDividendFinder dividend.Finder
}

func NewListStock(
	stockFinder stock.Finder,
	stockDividendFinder dividend.Finder,
) *listStocks {
	return &listStocks{
		stockFinder:         stockFinder,
		stockDividendFinder: stockDividendFinder,
	}
}

func (h *listStocks) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	exchange := command.(*appCommand.ListStocks).Exchange

	var (
		stks  []*stock.Stock
		rstks []*render.StockOutput
	)

	if exchange == "" {
		stks, err = h.stockFinder.FindAll()
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while finding stocks -> error [%s]",
				err,
			)

			return nil, err
		}

	} else {
		exchanges := strings.Split(exchange, ",")

		stks, err = h.stockFinder.FindAllByExchanges(exchanges)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while finding finding stocks from exchange [%s] -> error [%s]",
				exchange,
				err,
			)

			return nil, err
		}
	}

	for _, stk := range stks {
		var exDate time.Time

		d, err := h.stockDividendFinder.FindUpcoming(stk.ID)
		if err != nil {
			if err != mm.ErrNotFound {
				logger.FromContext(ctx).Errorf(
					"An error happen while finding next future dividend from [%s] -> error [%s]",
					stk.Symbol,
					err,
				)

				return nil, err
			}
		}

		if err == nil {
			exDate = d.ExDate
		}

		rstks = append(rstks, &render.StockOutput{
			Stock:          stk.Name,
			Market:         stk.Exchange.Symbol,
			Symbol:         stk.Symbol,
			Value:          stk.Value,
			High52Week:     stk.High52Week,
			Low52Week:      stk.Low52Week,
			BuyUnder:       stk.BuyUnder(),
			ExDate:         exDate,
			Dividend:       d.Amount,
			DividendStatus: d.Status,
			DYield:         stk.DividendYield,
			EPS:            stk.EPS,
			Change:         stk.Change,
			UpdatedAt:      stk.LastPriceUpdate,

			PriceWithHighLow: stk.ComparePriceWithHighLow(),
		})
	}

	return rstks, err
}
