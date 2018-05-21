package _import

import (
	"context"
	"time"

	"io"

	"strconv"

	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
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

		s, err := i.stockService.FindStockByName(line[2])
		if err != nil {
			return err
		}

		date := i.parseDateString(line[1])
		operation := account.Operation(line[3])
		amount, _ := strconv.Atoi(line[4])

		pf, _ := strconv.ParseFloat(line[5], 64)
		price := mm.Value{Amount: pf}

		pf, _ = strconv.ParseFloat(line[6], 64)
		priceChange := mm.Value{Amount: pf}

		pf, _ = strconv.ParseFloat(line[7], 64)
		priceChangeCommission := mm.Value{Amount: pf}

		pf, _ = strconv.ParseFloat(line[8], 64)
		value := mm.Value{Amount: pf}

		pf, _ = strconv.ParseFloat(line[9], 64)
		commission := mm.Value{Amount: pf}

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
