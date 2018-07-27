package render

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"

	"github.com/dohernandez/market-manager/pkg/application/util"
)

type screenListStocks struct {
	precision int
}

func NewScreenListStocks(precision int) *screenListStocks {
	return &screenListStocks{
		precision: precision,
	}
}

func (s *screenListStocks) Render(output interface{}) {
	sOutput := output.([]*StockOutput)

	noColor := color.New(color.Reset).FprintlnFunc()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	noColor(tw, "")
	s.renderStocks(tw, sOutput)
	noColor(tw, "")

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
			"%d\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %.*f\t %.*f\t %s\t %s\t %s\t",
			i+1,
			stk.Stock,
			stk.Market,
			stk.Symbol,
			util.SPrintValue(stk.Value, s.precision),
			util.SPrintValue(stk.High52Week, s.precision),
			util.SPrintValue(stk.Low52Week, s.precision),
			util.SPrintValue(stk.BuyUnder, s.precision),
			s.precision,
			stk.DYield,
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
