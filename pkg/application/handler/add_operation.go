package handler

import (
	"context"
	"fmt"

	"github.com/gogolfing/cbus"
	"github.com/pkg/errors"

	"time"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	addOperation struct {
		stockFinder stock.Finder
	}
)

func NewAddOperation(
	stockFinder stock.Finder,
) *addOperation {
	return &addOperation{
		stockFinder: stockFinder,
	}
}

func (h *addOperation) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	var (
		symbol                string
		action                operation.Action
		date                  time.Time
		price                 mm.Value
		priceChange           mm.Value
		priceChangeCommission mm.Value
		value                 mm.Value
		commission            mm.Value
		amount                int
	)
	switch cmd := command.(type) {
	case *appCommand.AddDividendOperation:
		action = operation.Dividend
		symbol = cmd.Stock
		date = parseOperationDateString(cmd.Date)
		value = mm.Value{Amount: cmd.Value, Currency: mm.Euro}
	case *appCommand.AddBuyOperation:
		action = operation.Buy
		symbol = cmd.Stock
		date = parseOperationDateString(cmd.Date)
		price = mm.Value{Amount: cmd.Price}
		priceChange = mm.Value{Amount: cmd.PriceChange}
		priceChangeCommission = mm.Value{Amount: cmd.PriceChangeCommission, Currency: mm.Euro}
		value = mm.Value{Amount: cmd.Value, Currency: mm.Euro}
		commission = mm.Value{Amount: cmd.Commission, Currency: mm.Euro}

		amount = cmd.Amount
	case *appCommand.AddSellOperation:
		action = operation.Sell
		symbol = cmd.Stock
		date = parseOperationDateString(cmd.Date)
		price = mm.Value{Amount: cmd.Price}
		priceChange = mm.Value{Amount: cmd.PriceChange}
		priceChangeCommission = mm.Value{Amount: cmd.PriceChangeCommission, Currency: mm.Euro}
		value = mm.Value{Amount: cmd.Value, Currency: mm.Euro}
		commission = mm.Value{Amount: cmd.Commission, Currency: mm.Euro}

		amount = cmd.Amount
	case *appCommand.AddInterestOperation:
		action = operation.Interest
		date = parseOperationDateString(cmd.Date)
		value = mm.Value{Amount: cmd.Value, Currency: mm.Euro}
	default:
		logger.FromContext(ctx).Error(
			"addOperation: Operation action not supported",
		)

		return nil, errors.New("operation action not supported")
	}

	s := new(stock.Stock)

	if action != operation.Interest {
		if symbol == "" {
			logger.FromContext(ctx).Error(
				"An error happen stock symbol not defined",
			)

			return nil, errors.New(fmt.Sprintf("find stock %s: %s", symbol, err.Error()))
		}

		s, err = h.stockFinder.FindBySymbol(symbol)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while finding stock by symbol [%s] -> error [%s]",
				symbol,
				err,
			)

			return nil, errors.New(fmt.Sprintf("find stock %s: %s", symbol, err.Error()))
		}
	}

	o := operation.NewOperation(date, s, action, amount, price, priceChange, priceChangeCommission, value, commission)

	return []*operation.Operation{
		o,
	}, nil
}
