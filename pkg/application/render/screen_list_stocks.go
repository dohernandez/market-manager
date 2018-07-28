package render

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"

	"sort"

	"time"

	"github.com/dohernandez/market-manager/pkg/application/util"
)

const (
	SortName   util.SortBy = "name"
	SortDyield util.SortBy = "dyield"
	SortExdate util.SortBy = "exdate"

	ExDateMonth util.GroupBy = "exdate"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// START Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type Stocks []*StockOutput

func (s Stocks) Len() int      { return len(s) }
func (s Stocks) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// StocksByName implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type StocksByName struct {
	Stocks
}

func (s StocksByName) Less(i, j int) bool {
	return s.Stocks[i].Stock < s.Stocks[j].Stock
}

// StocksByDividendYield implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type StocksByDividendYield struct {
	Stocks
}

func (s StocksByDividendYield) Less(i, j int) bool {
	return s.Stocks[i].DYield < s.Stocks[j].DYield
}

// StocksByDividendYield implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type StocksByExDate struct {
	Stocks
}

func (s StocksByExDate) Less(i, j int) bool {
	return s.Stocks[i].ExDate.Before(s.Stocks[j].ExDate)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// END Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	OutputScreenListStocks struct {
		Stocks []*StockOutput

		GroupBy util.GroupBy
		Sorting util.Sorting
	}

	screenListStocks struct {
		precision int
	}
)

func NewScreenListStocks(precision int) *screenListStocks {
	return &screenListStocks{
		precision: precision,
	}
}

func (s *screenListStocks) Render(output interface{}) {
	sOutput := output.(*OutputScreenListStocks)

	rstks := sOutput.Stocks
	switch sOutput.Sorting.By {
	case SortName:
		sort.Sort(StocksByName{rstks})
	case SortExdate:
		sort.Sort(StocksByExDate{rstks})
	case SortDyield:
		sort.Sort(StocksByDividendYield{rstks})
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	noColor := color.New(color.Reset).FprintlnFunc()

	switch sOutput.GroupBy {
	case ExDateMonth:
		mrstks := map[string][]*StockOutput{}

		for _, stk := range rstks {
			var key string

			if stk.ExDate.IsZero() {
				key = ""
			} else {
				exDate := stk.ExDate
				key = fmt.Sprintf("%s %d", exDate.Month(), exDate.Year())
			}

			mrstks[key] = append(mrstks[key], stk)
		}

		var keys []string
		for key := range mrstks {
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

		noColor(tw, "")
		for _, key := range keys {
			if key != "" {
				noColor(tw, "# ", key)
				noColor(tw, "")
			}

			s.renderStocks(tw, mrstks[key])
			noColor(tw, "")
		}
	default:
		noColor(tw, "")
		s.renderStocks(tw, rstks)
		noColor(tw, "")
	}

	tw.Flush()
}

func (s *screenListStocks) renderStocks(tw *tabwriter.Writer, stks []*StockOutput) {
	header := color.New(color.FgWhite).FprintlnFunc()

	header(tw, "#\t Stock\t Market\t Symbol\t Value\t High 52wk\t Low 52wk\t Buy Under\t D. Yield\t EPS\t Ex Date\t Change\t Updated At\t")

	normal := color.New(color.FgWhite).FprintlnFunc()
	overSell := color.New(color.FgGreen).FprintlnFunc()
	overBuy := color.New(color.FgRed).FprintlnFunc()

	for i, stk := range stks {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %.*f\t %s\t %s\t %s\t",
			i+1,
			stk.Stock,
			stk.Market,
			stk.Symbol,
			util.SPrintValue(stk.Value, s.precision),
			util.SPrintValue(stk.High52Week, s.precision),
			util.SPrintValue(stk.Low52Week, s.precision),
			util.SPrintValue(stk.BuyUnder, s.precision),
			util.SPrintPercentage(stk.DYield, s.precision),
			s.precision,
			stk.EPS,
			util.SPrintDate(stk.ExDate),
			util.SPrintValue(stk.Change, s.precision),
			util.SPrintDate(stk.UpdatedAt),
		)

		switch stk.PriceWithHighLow {
		case 1:
			overBuy(tw, str)
		case -1:
			overSell(tw, str)
		default:
			normal(tw, str)
		}
	}
}
