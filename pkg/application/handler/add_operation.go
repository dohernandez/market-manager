package handler

import (
	"context"

	"github.com/gogolfing/cbus"
	"github.com/pkg/errors"

	"time"

	"fmt"

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
	case *appCommand.AddDividend:
		action = operation.Dividend
		symbol = cmd.Stock
		date = h.parseDateString(cmd.Date)
		value = mm.Value{Amount: cmd.Value, Currency: mm.Euro}
	case *appCommand.AddBought:
		action = operation.Buy
		symbol = cmd.Stock
		date = h.parseDateString(cmd.Date)
		price = mm.Value{Amount: cmd.Price}
		priceChange = mm.Value{Amount: cmd.PriceChange}
		priceChangeCommission = mm.Value{Amount: cmd.PriceChangeCommission, Currency: mm.Euro}
		value = mm.Value{Amount: cmd.Value, Currency: mm.Euro}
		commission = mm.Value{Amount: cmd.Commission, Currency: mm.Euro}

		amount = cmd.Amount
	case *appCommand.AddSold:
		action = operation.Sell
		symbol = cmd.Stock
		date = h.parseDateString(cmd.Date)
		price = mm.Value{Amount: cmd.Price}
		priceChange = mm.Value{Amount: cmd.PriceChange}
		priceChangeCommission = mm.Value{Amount: cmd.PriceChangeCommission, Currency: mm.Euro}
		value = mm.Value{Amount: cmd.Value, Currency: mm.Euro}
		commission = mm.Value{Amount: cmd.Commission, Currency: mm.Euro}

		amount = cmd.Amount
	default:
		logger.FromContext(ctx).Error(
			"addOperation: Operation action not supported",
		)

		return nil, errors.New("operation action not supported")
	}

	if symbol == "" {
		logger.FromContext(ctx).Error(
			"An error happen stock symbol not defined",
		)

		return nil, errors.New(fmt.Sprintf("find stock %s: %s", symbol, err.Error()))
	}

	s, err := h.stockFinder.FindBySymbol(symbol)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while finding stock by symbol [%s] -> error [%s]",
			symbol,
			err,
		)

		return nil, errors.New(fmt.Sprintf("find stock %s: %s", symbol, err.Error()))
	}

	o := operation.NewOperation(date, s, action, amount, price, priceChange, priceChangeCommission, value, commission)

	return []*operation.Operation{
		o,
	}, nil
}

// parseDateString - parse a potentially partial date string to Time
func (h *addOperation) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}
