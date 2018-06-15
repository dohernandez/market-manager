package export_purchase

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"

	"github.com/dohernandez/market-manager/pkg/export"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// formatStocksToScreen - convert Items structure to csv string
func formatStocksToScreen(stks []*stock.Stock, sorting export.Sorting, gby export.GroupBy) *tabwriter.Writer {
	precision := 2

	sortStks := make([]*stock.Stock, 0, len(stks))
	for _, stk := range stks {
		sortStks = append(sortStks, stk)
	}

	switch sorting.By {
	case Name:
		sort.Sort(StocksByName{sortStks})
	case Exdate:
		sort.Sort(StocksByExDate{sortStks})
	case Dyield:
		sort.Sort(StocksByDividendYield{sortStks})
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	noColor := color.New(color.Reset).FprintlnFunc()
	header := color.New(color.FgWhite).FprintlnFunc()

	if gby != ExDateMonth {
		noColor(tw, "")
		header(tw, "#\t Stock\t Market\t Symbol\t Value\t High 52wk\t Low 52wk\t Buy Under\t D. Yield\t Ex Date\t Change\t Last Price Update\t")
		addStocksToTabWriter(tw, sortStks, precision)
	} else {
		mSortStks := map[string][]*stock.Stock{}

		for _, stk := range sortStks {
			var key string

			if len(stk.Dividends) == 0 {
				key = ""
			} else {
				exDate := stk.Dividends[0].ExDate
				key = fmt.Sprintf("%s %d", exDate.Month(), exDate.Year())
			}

			mSortStks[key] = append(mSortStks[key], stk)
		}

		var keys []string
		for key := range mSortStks {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool {
			if keys[i] == "" {
				return false
			}

			if keys[j] == "" {
				return true
			}

			iDate, _ := time.Parse("January 2006", keys[i])
			jDate, _ := time.Parse("January 2006", keys[j])

			return iDate.Before(jDate)
		})

		for _, key := range keys {
			if key != "" {
				noColor(tw, "")
				header(tw, "# ", key)
			}

			noColor(tw, "")

			header(tw, "#\t Stock\t Market\t Symbol\t Value\t High 52wk\t Low 52wk\t Buy Under\t D. Yield\t Ex Date\t Change\t Last Price Update\t")
			addStocksToTabWriter(tw, mSortStks[key], precision)
		}
	}

	return tw
}

func addStocksToTabWriter(tw *tabwriter.Writer, stks []*stock.Stock, precision int) {
	normal := color.New(color.FgWhite).FprintlnFunc()
	overSell := color.New(color.FgGreen).FprintlnFunc()
	overBuy := color.New(color.FgRed).FprintlnFunc()

	for i, stk := range stks {
		var (
			strDYield string
			strExDate time.Time
		)

		if stk.DividendYield > 0 {
			strDYield = fmt.Sprintf("%.*f%%", precision, stk.DividendYield)
		}

		if len(stk.Dividends) > 0 {
			strExDate = stk.Dividends[0].ExDate
		}

		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t",
			i+1,
			stk.Name,
			stk.Exchange.Symbol,
			stk.Symbol,
			export.PrintValue(stk.Value, precision),
			export.PrintValue(stk.High52week, precision),
			export.PrintValue(stk.Low52week, precision),
			export.PrintValue(stk.BuyUnder(), precision),
			strDYield,
			export.PrintDate(strExDate),
			export.PrintValue(stk.Change, precision),
			export.PrintDate(stk.LastPriceUpdate),
		)

		switch stk.ComparePriceWithHighLow() {
		case 1:
			overBuy(tw, str)
		case -1:
			overSell(tw, str)
		default:
			normal(tw, str)
		}
	}
}

// formatStocksToScreen - convert Items structure to csv string
func formatStocksDividendsToScreen(stks []*stock.Stock, sorting export.Sorting) *tabwriter.Writer {
	precision := 2
	sortStks := make([]*stock.Stock, 0, len(stks))

	for _, stk := range stks {
		sortStks = append(sortStks, stk)
	}

	switch sorting.By {
	case Exdate:
		sort.Sort(StocksByExDate{sortStks})
	default:
		sort.Sort(StocksByDividendYield{sortStks})
	}

	emptyTime := time.Time{}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(w, "")

	header := color.New(color.Bold, color.FgBlack).FprintlnFunc()
	header(w, "#\t Stock\t Market\t Symbol\t Value\t D. Yield\t Ex Date\t Record Date\t Status\t Amount\t Change\t Last Price Update\t")

	pColor := color.New(color.Bold, color.FgWhite).FprintlnFunc()
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
				export.PrintDate(stk.LastPriceUpdate),
			)
			pColor(w, str)
		}
	}

	noColor(w, "")

	return w
}
