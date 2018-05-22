package _import

import (
	"context"
	"time"

	"io"

	"strconv"

	"strings"

	"fmt"

	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
	"github.com/pkg/errors"
)

type (
	ImportAccount struct {
		ctx    context.Context
		reader Reader

		stockService   *stock.Service
		accountService *account.Service
	}
)

func NewImportAccount(
	ctx context.Context,
	reader Reader,
	stockService *stock.Service,
	accountService *account.Service,
) *ImportAccount {
	return &ImportAccount{
		ctx:            ctx,
		reader:         reader,
		stockService:   stockService,
		accountService: accountService,
	}
}

func (i *ImportAccount) Import() error {
	i.reader.Open()
	defer i.reader.Close()

	var as []*account.Account

	for {
		line, err := i.reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(i.ctx).Fatal(err)
		}

		operation, err := i.parseOperationString(line[3])
		if err != nil {
			return err
		}

		s := &stock.Stock{}
		if operation != account.Connectivity && operation != account.Interest {
			s, err = i.stockService.FindStockByName(line[2])
			if err != nil {
				return errors.New(fmt.Sprintf("%s: %s", line[2], err.Error()))
			}
		}

		date := i.parseDateString(line[1])
		amount, _ := strconv.Atoi(line[4])

		price := mm.Value{Amount: i.parsePriceString(line[5])}
		priceChange := mm.Value{Amount: i.parsePriceString(line[6])}
		priceChangeCommission := mm.Value{Amount: i.parsePriceString(line[7])}
		value := mm.Value{Amount: i.parsePriceString(line[8])}
		commission := mm.Value{Amount: i.parsePriceString(line[9])}

		a := account.NewAccount(date, s, operation, amount, price, priceChange, priceChangeCommission, value, commission)

		as = append(as, a)
	}

	return i.accountService.SaveAll(as)
}

// parseDateString - parse a potentially partial date string to Time
func (i *ImportAccount) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parseOperationString - parse a potentially partial date string to Time
func (i *ImportAccount) parseOperationString(operation string) (account.Operation, error) {
	if operation == "" {
		return account.Operation(""), errors.New("operation can not be empty")
	}

	switch operation {
	case "Compra":
		return account.Buy, nil
	case "Venta":
		return account.Sell, nil
	case "Conectividad":
		return account.Connectivity, nil
	case "Dividendo":
		return account.Dividend, nil
	case "Interés":
		return account.Interest, nil
	}

	return account.Operation(""), errors.New("operation not valid")
}

// parseDateString - parse a potentially partial date string to Time
func (i *ImportAccount) parsePriceString(price string) float64 {
	price = strings.Replace(price, ",", ".", 1)

	p, _ := strconv.ParseFloat(price, 64)

	return p
}
