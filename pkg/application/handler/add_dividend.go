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
	addDividend struct {
		stockFinder stock.Finder
	}
)

func NewAddDividend(
	stockFinder stock.Finder,
) *addDividend {
	return &addDividend{
		stockFinder: stockFinder,
	}
}

func (h *addDividend) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	ad := command.(*appCommand.AddDividend)

	symbol := ad.Stock
	s, err := h.stockFinder.FindByName(symbol)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while finding stock by name [%s] -> error [%s]",
			symbol,
			err,
		)

		return nil, errors.New(fmt.Sprintf("find stock %s: %s", symbol, err.Error()))
	}

	dDate := h.parseDateString(ad.Date)
	//amount, _ := strconv.Atoi(ad.Amount)
	dValue := ad.Value

	price := mm.Value{Amount: 0}
	priceChange := mm.Value{Amount: 0}
	priceChangeCommission := mm.Value{Amount: 0, Currency: mm.Euro}
	value := mm.Value{Amount: dValue, Currency: mm.Euro}
	commission := mm.Value{Amount: 0, Currency: mm.Euro}

	o := operation.NewOperation(dDate, s, operation.Dividend, 0, price, priceChange, priceChangeCommission, value, commission)

	return []*operation.Operation{
		o,
	}, nil
}

// parseDateString - parse a potentially partial date string to Time
func (h *addDividend) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}
