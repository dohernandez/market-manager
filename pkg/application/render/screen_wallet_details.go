package render

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"

	"time"

	"sort"

	"github.com/dohernandez/market-manager/pkg/application/util"
)

const (
	WalletStock    util.SortBy = "stock"
	WalletInvested util.SortBy = "invested"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// START Wallet Items Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type WalletItems []*WalletStockOutput

func (s WalletItems) Len() int      { return len(s) }
func (s WalletItems) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// WalletItemsByName implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded ExportWalletItems value.
type WalletItemsByName struct {
	WalletItems
}

func (s WalletItemsByName) Less(i, j int) bool {
	return s.WalletItems[i].Stock < s.WalletItems[j].Stock
}

// WalletItemsByInvested implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded ExportWalletItems value.
type WalletItemsByInvested struct {
	WalletItems
}

func (s WalletItemsByInvested) Less(i, j int) bool {
	return s.WalletItems[i].Invested.Amount < s.WalletItems[j].Invested.Amount
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// END Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	OutputScreenWalletDetails struct {
		WalletDetails WalletDetailsOutput

		GroupBy util.GroupBy
		Sorting util.Sorting
	}

	screenWalletDetails struct {
		precision int
	}
)

func NewScreenWalletDetails(precision int) *screenWalletDetails {
	return &screenWalletDetails{
		precision: precision,
	}
}

func (s *screenWalletDetails) Render(output interface{}) {
	sOutput := output.(*OutputScreenWalletDetails)

	walletStockOutputs := sOutput.WalletDetails.WalletStockOutputs
	walletOutput := sOutput.WalletDetails.WalletOutput

	switch sOutput.Sorting.By {
	case WalletInvested:
		sort.Sort(WalletItemsByInvested{walletStockOutputs})
	default:
		sort.Sort(WalletItemsByName{walletStockOutputs})
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	noColor := color.New(color.Reset).FprintlnFunc()

	noColor(tw, "")
	s.renderGeneral(tw, walletOutput)
	noColor(tw, "")
	s.renderItemStocks(tw, walletStockOutputs)
	noColor(tw, "")
	s.renderWalletDividendProjected(tw, walletOutput)
	noColor(tw, "")
	s.renderStocksDividends(tw, walletStockOutputs)
	noColor(tw, "")
	s.renderStocks(tw, walletStockOutputs)
	noColor(tw, "")

	tw.Flush()
}

func (s *screenWalletDetails) renderGeneral(tw *tabwriter.Writer, wOutput WalletOutput) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# General")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "Invested\t Capital\t Funds\t Free Margin\t Net Capital\t Net Benefits\t % Benefits\t Dividends\t D. Yield\t Connection\t Interest\t Commissions\t")

	pColor := color.New(color.FgGreen).FprintlnFunc()
	if wOutput.PercentageBenefits < 0 {
		pColor = color.New(color.FgRed).FprintlnFunc()
	}

	str := fmt.Sprintf(
		"%s\t %s\t %s\t %s\t %s\t %s\t %.*f%%\t %s\t %s\t %s\t %s\t %s\t",
		util.SPrintValue(wOutput.Invested, s.precision),
		util.SPrintValue(wOutput.Capital, s.precision),
		util.SPrintValue(wOutput.Funds, s.precision),
		util.SPrintValue(wOutput.FreeMargin, s.precision),
		util.SPrintValue(wOutput.NetCapital, s.precision),
		util.SPrintValue(wOutput.NetBenefits, s.precision),
		s.precision,
		wOutput.PercentageBenefits,
		util.SPrintValue(wOutput.DividendPayed, s.precision),
		util.SPrintPercentage(wOutput.DividendYearYield, s.precision),
		util.SPrintValue(wOutput.Connection, s.precision),
		util.SPrintValue(wOutput.Interest, s.precision),
		util.SPrintValue(wOutput.Commission, s.precision),
	)

	pColor(tw, str)
}

func (s *screenWalletDetails) renderItemStocks(tw *tabwriter.Writer, wStocks []*WalletStockOutput) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Stocks")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Capital\t Invested\t % \t Dividend\t Buys\t Sells\t Benefits\t % \t Change\t")

	inProfits := color.New(color.FgGreen).FprintlnFunc()
	inLooses := color.New(color.FgRed).FprintlnFunc()

	for i, stk := range wStocks {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %.*f%%\t %s\t %s\t %s\t %s\t %.*f%%\t %s\t",
			i+1,
			stk.Stock,
			stk.Market,
			stk.Symbol,
			stk.Amount,
			util.SPrintValue(stk.Capital, s.precision),
			util.SPrintValue(stk.Invested, s.precision),
			s.precision,
			stk.PercentageWallet,
			util.SPrintValue(stk.Dividend, s.precision),
			util.SPrintValue(stk.Buys, s.precision),
			util.SPrintValue(stk.Sells, s.precision),
			util.SPrintValue(stk.NetBenefits, s.precision),
			s.precision,
			stk.PercentageBenefits,
			util.SPrintValue(stk.Change, s.precision),
		)

		if stk.PercentageBenefits > 0 {
			inProfits(tw, str)
		} else {
			inLooses(tw, str)
		}
	}
}

func (s *screenWalletDetails) renderWalletDividendProjected(tw *tabwriter.Writer, wOutput WalletOutput) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Dividend Projected")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "Month\t Dividend\t D. Yield\t          Year\t Dividend\t D. Yield\t")

	now := time.Now()

	inNormal := color.New(color.FgWhite).FprintlnFunc()
	inNormal(tw, fmt.Sprintf(
		"%s\t %s\t %s\t          %d\t %s\t %s\t",
		now.Month(),
		util.SPrintValue(wOutput.DividendMonthProjected, s.precision),
		util.SPrintPercentage(wOutput.DividendMonthYield, s.precision),
		now.Year(),
		util.SPrintValue(wOutput.DividendYearProjected, s.precision),
		util.SPrintPercentage(wOutput.DividendYearYield, s.precision),
	))
}

func (s *screenWalletDetails) renderStocksDividends(tw *tabwriter.Writer, wStocks []*WalletStockOutput) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Stocks Dividends")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Price\t WA Price\t Ex Date\t Dividend\t D. Yield\t WA D. Yield\t Updated At\t")

	inNormal := color.New(color.FgWhite).FprintlnFunc()
	inHeightLight := color.New(color.FgYellow).FprintlnFunc()

	now := time.Now()

	for i, stk := range wStocks {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s %s\t %s\t %s\t %s\t %s\t",
			i+1,
			stk.Stock,
			stk.Market,
			stk.Symbol,
			stk.Amount,
			util.SPrintValue(stk.Value, s.precision),
			util.SPrintValue(stk.WAPrice, s.precision),
			util.SPrintDate(stk.ExDate),
			util.SPrintInitialDividendStatus(stk.DividendStatus),
			util.SPrintValue(stk.Dividend, s.precision),
			util.SPrintPercentage(stk.DYield, s.precision),
			util.SPrintPercentage(stk.WADYield, s.precision),
			util.SPrintDateTime(stk.UpdatedAt),
		)

		if stk.ExDate.Month() == now.Month() && stk.ExDate.Year() == now.Year() {
			inHeightLight(tw, str)
		} else {
			inNormal(tw, str)
		}
	}
}

func (s *screenWalletDetails) renderStocks(tw *tabwriter.Writer, wStocks []*WalletStockOutput) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Stocks")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Price\t WA Price\t High 52wk\t Low 52wk\t Buy Under\t EPS\t Change\t Updated At\t")

	normal := color.New(color.FgWhite).FprintlnFunc()
	overSell := color.New(color.FgGreen).FprintlnFunc()
	overBuy := color.New(color.FgRed).FprintlnFunc()

	for i, stk := range wStocks {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s\t %s\t %s\t %.*f\t %s\t %s\t",
			i+1,
			stk.Stock,
			stk.Market,
			stk.Symbol,
			stk.Amount,
			util.SPrintValue(stk.Value, s.precision),
			util.SPrintValue(stk.WAPrice, s.precision),
			util.SPrintValue(stk.High52Week, s.precision),
			util.SPrintValue(stk.Low52Week, s.precision),
			util.SPrintValue(stk.BuyUnder, s.precision),
			s.precision,
			stk.EPS,
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
