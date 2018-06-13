package export_purchase

import (
	"context"

	"strings"

	"time"

	"strconv"

	"github.com/dohernandez/market-manager/pkg/export"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

const (
	Dyield export.SortBy = "dyield"
	Exdate export.SortBy = "exdate"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// START Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type Stocks []*stock.Stock

func (s Stocks) Len() int      { return len(s) }
func (s Stocks) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// StocksByName implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type StocksByName struct {
	Stocks
}

func (s StocksByName) Less(i, j int) bool {
	return s.Stocks[i].Name < s.Stocks[j].Name
}

// StocksByDividendYield implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type StocksByDividendYield struct {
	Stocks
}

func (s StocksByDividendYield) Less(i, j int) bool {
	return s.Stocks[i].DividendYield < s.Stocks[j].DividendYield
}

// StocksByDividendYield implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type StocksByExDate struct {
	Stocks
}

func (s StocksByExDate) Less(i, j int) bool {
	return s.Stocks[i].Dividends[len(s.Stocks[i].Dividends)-1].ExDate.Before(s.Stocks[j].Dividends[len(s.Stocks[j].Dividends)-1].ExDate)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// END Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	exportStock struct {
		ctx     context.Context
		sorting export.Sorting

		purchaseService *purchase.Service
	}
)

func NewExportStock(ctx context.Context, sorting export.Sorting, purchaseService *purchase.Service) *exportStock {
	return &exportStock{
		ctx:             ctx,
		sorting:         sorting,
		purchaseService: purchaseService,
	}
}

func (e *exportStock) Export() error {
	exchange := e.ctx.Value("exchange").(string)

	var (
		stks []*stock.Stock
		err  error
	)

	if exchange == "" {
		stks, err = e.retrieveStocks()
	} else {
		exchanges := strings.Split(exchange, ",")

		stks, err = e.purchaseService.StocksByExchanges(exchanges)
	}
	if err != nil {
		return err
	}

	tabw := formatStocksToScreen(stks)
	tabw.Flush()

	return nil
}

func (e *exportStock) retrieveStocks() ([]*stock.Stock, error) {
	var stks []*stock.Stock
	symbol := e.ctx.Value("symbol").(string)

	if symbol != "" {
		stk, err := e.purchaseService.FindStockBySymbol(symbol)
		if err != nil {
			if err == mm.ErrNotFound {
				return stks, nil
			}

			return nil, err
		}

		stks = append(stks, stk)
		return stks, nil
	}

	return e.purchaseService.Stocks()
}

type (
	exportStockWithDividends struct {
		ctx     context.Context
		sorting export.Sorting

		purchaseService *purchase.Service
	}
)

func NewExportStockWithDividends(ctx context.Context, sorting export.Sorting, purchaseService *purchase.Service) *exportStockWithDividends {
	return &exportStockWithDividends{
		ctx:             ctx,
		sorting:         sorting,
		purchaseService: purchaseService,
	}
}

func (e *exportStockWithDividends) Export() error {
	year := e.ctx.Value("year").(string)
	month := e.ctx.Value("month").(string)

	now := time.Now()
	if year == "" {
		year = strconv.Itoa(now.Year())
	}

	if month == "" {
		month = strconv.Itoa(int(now.Month()))
	}

	stks, err := e.purchaseService.StocksByDividendAnnounceProjectYearAndMonth(year, month)
	if err != nil {
		return err
	}

	tabw := formatStocksDividendsToScreen(stks, e.sorting)
	tabw.Flush()

	return nil
}
