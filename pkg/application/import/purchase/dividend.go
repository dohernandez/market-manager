package import_purchase

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/import"
	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	ImportStockDividend struct {
		ctx    context.Context
		reader _import.Reader

		purchaseService *service.Purchase
	}
)

func NewImportStockDividend(
	ctx context.Context,
	reader _import.Reader,
	purchaseService *service.Purchase,
) *ImportStockDividend {
	return &ImportStockDividend{
		ctx:             ctx,
		reader:          reader,
		purchaseService: purchaseService,
	}
}

func (i *ImportStockDividend) Import() error {
	i.reader.Open()
	defer i.reader.Close()

	symbol, ok := i.ctx.Value("stock").(string)
	if !ok {
		return errors.New("Stock symbol not defined")
	}

	stk, err := i.purchaseService.FindStockBySymbol(symbol)
	if err != nil {
		return errors.New(fmt.Sprintf("%s [symbol %s]", err.Error(), symbol))
	}

	var ds []dividend.StockDividend

	for {
		line, err := i.reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(i.ctx).Fatal(err)
		}

		var status dividend.Status
		switch line[3] {
		case "Payed":
			status = dividend.Payed
		case "Announced":
			status = dividend.Announced
		case "Projected":
			status = dividend.Projected
		default:
			return errors.New("Dividend status not defined")
		}

		d := dividend.StockDividend{
			Status: status,
		}

		if len(line[0]) > 0 {
			d.ExDate = i.parseDateString(line[0])
		}

		if len(line[1]) > 0 {
			d.PaymentDate = i.parseDateString(line[1])
		}

		if len(line[2]) > 0 {
			d.RecordDate = i.parseDateString(line[2])
		}

		if len(line[4]) > 0 {
			d.Amount = mm.ValueDollarFromString(line[4])
		}

		if len(line[5]) > 0 {
			cfp, _ := strconv.ParseFloat(line[5], 64)
			d.ChangeFromPrev = cfp
		}

		if len(line[6]) > 0 {
			cfpy, _ := strconv.ParseFloat(line[6], 64)
			d.ChangeFromPrevYear = cfpy
		}

		if len(line[7]) > 0 {
			p12my, _ := strconv.ParseFloat(line[7], 64)
			d.Prior12MonthsYield = p12my
		}

		ds = append(ds, d)
	}

	stk.Dividends = ds

	return i.purchaseService.UpdateStockDividends(stk)
}

// parseDateString - parse a potentially partial date string to Time
func (i *ImportStockDividend) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2-Jan-2006", dt)

	return t
}
