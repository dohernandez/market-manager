package import_purchase

import (
	"context"
	"io"
	"strconv"
	"time"

	"fmt"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	ImportStockDividend struct {
		ctx    context.Context
		reader _import.Reader

		purchaseService *purchase.Service
	}
)

func NewImportStockDividend(
	ctx context.Context,
	reader _import.Reader,
	purchaseService *purchase.Service,
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

	status := dividend.Payed

	if s, _ := i.ctx.Value("status").(string); s == "projected" {
		status = dividend.Projected
	}

	for {
		line, err := i.reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(i.ctx).Fatal(err)
		}

		if status == dividend.Payed {
			a, _ := strconv.ParseFloat(line[3], 64)
			cfp, _ := strconv.ParseFloat(line[4], 64)
			cfpy, _ := strconv.ParseFloat(line[5], 64)
			p12my, _ := strconv.ParseFloat(line[6], 64)

			ds = append(ds, dividend.StockDividend{
				ExDate:             i.parseDateString(line[0]),
				PaymentDate:        i.parseDateString(line[1]),
				RecordDate:         i.parseDateString(line[2]),
				Status:             status,
				Amount:             a,
				ChangeFromPrev:     cfp,
				ChangeFromPrevYear: cfpy,
				Prior12MonthsYield: p12my,
			})
		} else {
			a, _ := strconv.ParseFloat(line[2], 64)

			ds = append(ds, dividend.StockDividend{
				ExDate:      i.parseDateString(line[0]),
				PaymentDate: i.parseDateString(line[1]),
				Status:      status,
				Amount:      a,
			})
		}

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
