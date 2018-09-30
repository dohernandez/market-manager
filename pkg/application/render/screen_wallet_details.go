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

		Precision int
	}

	screenWalletDetails struct {
	}
)

func NewScreenWalletDetails() *screenWalletDetails {
	return &screenWalletDetails{}
}

func (s *screenWalletDetails) Render(output interface{}) {
	sOutput := output.(*OutputScreenWalletDetails)

	walletStockOutputs := sOutput.WalletDetails.WalletStockOutputs
	walletOutput := sOutput.WalletDetails.WalletOutput
	precision := sOutput.Precision

	switch sOutput.Sorting.By {
	case WalletInvested:
		sort.Sort(WalletItemsByInvested{walletStockOutputs})
	default:
		sort.Sort(WalletItemsByName{walletStockOutputs})
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	noColor := color.New(color.Reset).FprintlnFunc()

	noColor(tw, "")
	s.renderGeneral(tw, walletOutput, precision)
	noColor(tw, "")
	s.renderItemStocks(tw, walletStockOutputs, precision)
	noColor(tw, "")
	s.renderWalletDividendProjected(tw, walletOutput, precision)
	noColor(tw, "")
	s.renderStocksDividends(tw, walletStockOutputs, precision)
	noColor(tw, "")
	s.renderStocks(tw, walletStockOutputs, precision)
	noColor(tw, "")

	tw.Flush()
}

func (s *screenWalletDetails) renderGeneral(tw *tabwriter.Writer, wOutput WalletOutput, precision int) {
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
		util.SPrintValue(wOutput.Invested, precision),
		util.SPrintValue(wOutput.Capital, precision),
		util.SPrintValue(wOutput.Funds, precision),
		util.SPrintValue(wOutput.FreeMargin, precision),
		util.SPrintValue(wOutput.NetCapital, precision),
		util.SPrintValue(wOutput.NetBenefits, precision),
		precision,
		wOutput.PercentageBenefits,
		util.SPrintValue(wOutput.DividendPayed, precision),
		util.SPrintPercentage(wOutput.DividendPayedYield, precision),
		util.SPrintValue(wOutput.Connection, precision),
		util.SPrintValue(wOutput.Interest, precision),
		util.SPrintValue(wOutput.Commission, precision),
	)

	pColor(tw, str)
}

func (s *screenWalletDetails) renderItemStocks(tw *tabwriter.Writer, wStocks []*WalletStockOutput, precision int) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Stocks")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t AMT\t Capital\t Invested\t % \t Dividend\t Buys\t Sells\t Benefits\t % \t Change\t")

	inProfits := color.New(color.FgGreen).FprintlnFunc()
	inLooses := color.New(color.FgRed).FprintlnFunc()

	for i, stk := range wStocks {

		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t %.*f%%\t %s\t",
			i+1,
			util.SPrintTruncate(stk.Stock, 27),
			stk.Market,
			stk.Symbol,
			stk.Amount,
			util.SPrintValue(stk.Capital, precision),
			util.SPrintValue(stk.Invested, precision),
			util.SPrintPercentage(stk.PercentageWallet, precision),
			util.SPrintValue(stk.DividendPayed, precision),
			util.SPrintValue(stk.Buys, precision),
			util.SPrintValue(stk.Sells, precision),
			util.SPrintValue(stk.NetBenefits, precision),
			precision,
			stk.PercentageBenefits,
			util.SPrintValue(stk.Change, precision),
		)

		if stk.PercentageBenefits > 0 {
			inProfits(tw, str)
		} else {
			inLooses(tw, str)
		}
	}
}

func (s *screenWalletDetails) renderWalletDividendProjected(tw *tabwriter.Writer, wOutput WalletOutput, precision int) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Dividend Projected")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "Month\t Dividend\t D. Yield\t          Month\t Dividend\t D. Yield\t          Month\t Dividend\t D. Yield\t          Year\t Dividend\t D. Yield\t")

	now := time.Now()

	inNormal := color.New(color.FgWhite).FprintlnFunc()
	inNormal(tw, fmt.Sprintf(
		"%s\t %s\t %s\t          %s\t %s\t %s\t          %s\t %s\t %s\t          %d\t %s\t %s\t",
		wOutput.DividendProjected[0].Month,
		util.SPrintValue(wOutput.DividendProjected[0].Projected, precision),
		util.SPrintPercentage(wOutput.DividendProjected[0].Yield, precision),
		wOutput.DividendProjected[1].Month,
		util.SPrintValue(wOutput.DividendProjected[1].Projected, precision),
		util.SPrintPercentage(wOutput.DividendProjected[1].Yield, precision),
		wOutput.DividendProjected[2].Month,
		util.SPrintValue(wOutput.DividendProjected[2].Projected, precision),
		util.SPrintPercentage(wOutput.DividendProjected[2].Yield, precision),
		now.Year(),
		util.SPrintValue(wOutput.DividendYearProjected, precision),
		util.SPrintPercentage(wOutput.DividendYearYield, precision),
	))
}

func (s *screenWalletDetails) renderStocksDividends(tw *tabwriter.Writer, wStocks []*WalletStockOutput, precision int) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Stocks Dividends")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t AMT\t Price\t WA Price\t Ex Date\t Dividend\t Retention (%)\t D. Yield\t WA D. Yield\t D. Pay\t")

	inNormal := color.New(color.FgWhite).FprintlnFunc()
	inHeightLight := color.New(color.FgYellow).FprintlnFunc()

	now := time.Now()

	for i, stk := range wStocks {
		var strPercentageRetention string

		if stk.DividendRetention.Amount > 0 {
			strPercentageRetention = fmt.Sprintf("(%.2f%%)", stk.PercentageRetention)
		}

		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s %s\t %s\t %s %s\t %s\t %s\t %s\t",
			i+1,
			util.SPrintTruncate(stk.Stock, 27),
			stk.Market,
			stk.Symbol,
			stk.Amount,
			util.SPrintValue(stk.Value, precision),
			util.SPrintValue(stk.WAPrice, precision),
			util.SPrintDate(stk.ExDate),
			util.SPrintInitialDividendStatus(stk.DividendStatus),
			util.SPrintValue(stk.Dividend, precision),
			util.SPrintValue(stk.DividendRetention, 4),
			strPercentageRetention,
			util.SPrintPercentage(stk.DYield, precision),
			util.SPrintPercentage(stk.WADYield, precision),
			util.SPrintValue(stk.DividendToPay, precision),
		)

		if stk.ExDate.Month() == now.Month() && stk.ExDate.Year() == now.Year() {
			inHeightLight(tw, str)
		} else {
			inNormal(tw, str)
		}
	}
}

func (s *screenWalletDetails) renderStocks(tw *tabwriter.Writer, wStocks []*WalletStockOutput, precision int) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Stocks Price")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t AMT\t Price\t WA Price\t High 52wk\t Low 52wk\t HV 52wk\t HV 20day\t Buy Under\t EPS\t PER\t Change\t")

	normal := color.New(color.FgWhite).FprintlnFunc()
	overSell := color.New(color.FgGreen).FprintlnFunc()
	overBuy := color.New(color.FgRed).FprintlnFunc()

	for i, stk := range wStocks {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s\t %s\t %.*f\t %.*f\t %s\t %.*f\t %.*f\t %s\t",
			i+1,
			util.SPrintTruncate(stk.Stock, 27),
			stk.Market,
			stk.Symbol,
			stk.Amount,
			util.SPrintValue(stk.Value, precision),
			util.SPrintValue(stk.WAPrice, precision),
			util.SPrintValue(stk.High52Week, precision),
			util.SPrintValue(stk.Low52Week, precision),
			precision,
			stk.HV52Week,
			precision,
			stk.HV20Day,
			util.SPrintValue(stk.BuyUnder, precision),
			precision,
			stk.EPS,
			precision,
			stk.PER,
			util.SPrintValue(stk.Change, precision),
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
