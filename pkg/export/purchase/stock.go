package export_purchase

import (
	"context"

	"strings"

	"github.com/dohernandez/market-manager/pkg/export"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

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
		stks, err = e.purchaseService.Stocks()
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// START Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ByName implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type Stocks []*stock.Stock

func (s Stocks) Len() int      { return len(s) }
func (s Stocks) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type StocksByName struct {
	Stocks
}

func (s StocksByName) Less(i, j int) bool {
	return s.Stocks[i].Name < s.Stocks[j].Name
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// END Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
