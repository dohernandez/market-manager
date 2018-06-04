package export_purchase

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// formatStocksToScreen - convert Items structure to csv string
func formatStocksToScreen(stks []*stock.Stock) *tabwriter.Writer {
	precision := 2
	sortStks := make([]*stock.Stock, 0, len(stks))

	for _, stk := range stks {
		sortStks = append(sortStks, stk)
	}

	sort.Sort(StocksByName{sortStks})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	fmt.Fprintln(w, "#\t Stock\t Market\t Symbol\t Value\t Dividend Yield\t Change\t Last Price Update\t")
	for i, stk := range sortStks {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %.*f\t %.*f%%\t %.*f\t %s\t",
			i+1,
			stk.Name,
			stk.Exchange.Symbol,
			stk.Symbol,
			precision,
			stk.Value.Amount,
			precision,
			stk.DividendYield,
			precision,
			stk.Change.Amount,
			stk.LastPriceUpdate.Format(time.RFC822),
		)
		fmt.Fprintln(w, str)
	}

	return w
}

// formatStocksToScreen - convert Items structure to csv string
func formatStocksDividendsToScreen(stks []*stock.Stock) *tabwriter.Writer {
	precision := 2
	sortStks := make([]*stock.Stock, 0, len(stks))

	for _, stk := range stks {
		sortStks = append(sortStks, stk)
	}

	sort.Sort(StocksByDividendYield{sortStks})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	emptyTime := time.Time{}
	fmt.Fprintln(w, "#\t Stock\t Market\t Symbol\t Value\t D. Yield\t Ex Date\t Record Date\t Status\t Amount\t Change\t Last Price Update\t")
	for i, stk := range sortStks {
		for _, d := range stk.Dividends {
			record := d.RecordDate.Format("Mon Jan 2 06")
			if emptyTime.Equal(d.RecordDate) {
				record = ""
			}

			str := fmt.Sprintf(
				"%d\t %s\t %s\t %s\t %.*f\t %.*f%%\t %s\t %s\t %s\t %.*f\t %.*f\t %s\t",
				i+1,
				stk.Name,
				stk.Exchange.Symbol,
				stk.Symbol,
				precision,
				stk.Value.Amount,
				precision,
				stk.DividendYield,
				d.ExDate.Format("Mon Jan 2 06"),
				record,
				d.Status,
				precision,
				d.Amount.Amount,
				precision,
				stk.Change.Amount,
				stk.LastPriceUpdate.Format("02 Jan 06 15:04"),
			)
			fmt.Fprintln(w, str)
		}
	}

	return w
}
