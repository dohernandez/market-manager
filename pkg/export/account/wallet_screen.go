package export_account

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/fatih/color"

	"time"

	"github.com/dohernandez/market-manager/pkg/export"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

// formatWalletItemsToScreen - convert Items structure to screen
func formatWalletItemsToScreen(w *wallet.Wallet, sorting export.Sorting) *tabwriter.Writer {
	precision := 2
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	formatWalletToScreen(tw, precision, w)

	sortItems := make([]*wallet.Item, 0, len(w.Items))
	for _, item := range w.Items {
		sortItems = append(sortItems, item)
	}

	switch sorting.By {
	case Invested:
		sort.Sort(WalletItemsByInvested{sortItems})
	default:
		sort.Sort(WalletItemsByName{sortItems})
	}

	formatItemsToScreen(tw, precision, w, sortItems)
	formatWalletDividendProjected(tw, precision, w)
	formatStockItemsDividendToScreen(tw, precision, sortItems)
	formatStockItemsToScreen(tw, precision, sortItems)

	fmt.Fprintln(tw, "")

	return tw
}

// formatWalletToScreen - convert wallet structure to screen
func formatWalletToScreen(tw *tabwriter.Writer, precision int, w *wallet.Wallet) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")
	noColor(tw, "# General")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "Invested\t Capital\t Funds\t Free Margin\t Net Capital\t Net Benefits\t % Benefits\t Dividends\t D. Yield\t Connection\t Interest\t Commissions\t")

	pColor := color.New(color.FgGreen).FprintlnFunc()
	if w.PercentageBenefits() < 0 {
		pColor = color.New(color.FgRed).FprintlnFunc()
	}

	wDYield := w.Dividend.Amount * 4 / w.Invested.Amount * 100

	str := fmt.Sprintf(
		"%s\t %s\t %s\t %s\t %s\t %s\t %.*f%%\t %s\t %.*f%%\t %s\t %s\t %s\t",
		export.PrintValue(w.Invested, precision),
		export.PrintValue(w.Capital, precision),
		export.PrintValue(w.Funds, precision),
		export.PrintValue(w.FreeMargin(), precision),
		export.PrintValue(w.NetCapital(), precision),
		export.PrintValue(w.NetBenefits(), precision),
		precision,
		w.PercentageBenefits(),
		export.PrintValue(w.Dividend, precision),
		precision,
		wDYield,
		export.PrintValue(w.Connection, precision),
		export.PrintValue(w.Interest, precision),
		export.PrintValue(w.Commission, precision),
	)
	pColor(tw, str)
}

// formatItemsToScreen - convert Items structure to screen
func formatItemsToScreen(tw *tabwriter.Writer, precision int, w *wallet.Wallet, items []*wallet.Item) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Capital\t Invested\t % \t Dividend\t Buys\t Sells\t Benefits\t % \t Change\t")

	inProfits := color.New(color.FgGreen).FprintlnFunc()
	inLooses := color.New(color.FgRed).FprintlnFunc()

	for i, item := range items {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %.*f%%\t %s\t %s\t %s\t %s\t %.*f%%\t %s\t",
			i+1,
			item.Stock.Name,
			item.Stock.Exchange.Symbol,
			item.Stock.Symbol,
			item.Amount,
			export.PrintValue(item.Capital(), precision),
			export.PrintValue(item.Invested, precision),
			precision,
			item.PercentageInvestedRepresented(w.Capital.Amount),
			export.PrintValue(item.Dividend, precision),
			export.PrintValue(item.Buys, precision),
			export.PrintValue(item.Sells, precision),
			export.PrintValue(item.NetBenefits(), precision),
			precision,
			item.PercentageBenefits(),
			export.PrintValue(item.Change(), precision),
		)

		if item.PercentageBenefits() > 0 {
			inProfits(tw, str)
		} else {
			inLooses(tw, str)
		}
	}
}

// formatWalletDividendProjected ...
func formatWalletDividendProjected(tw *tabwriter.Writer, precision int, w *wallet.Wallet) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")
	noColor(tw, "# Dividend")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "Dividend Projected\t Retention\t Final\t")

	dProjected := w.DividendProjectedNextMonth()

	var retention mm.Value
	retention.Amount = 15 * dProjected.Amount / 100
	retention.Currency = dProjected.Currency

	inNormal := color.New(color.FgWhite).FprintlnFunc()
	inNormal(tw, fmt.Sprintf(
		"%s\t %s\t %s\t",
		export.PrintValue(dProjected, precision),
		export.PrintValue(retention, precision),
		export.PrintValue(dProjected.Decrease(retention), precision),
	))
}

// formatStockItemsDividendToScreen ...
func formatStockItemsDividendToScreen(tw *tabwriter.Writer, precision int, items []*wallet.Item) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Price\t WA Price\t Ex Date\t Dividend\t D. Yield\t WA D. Yield\t Last Price Update\t")

	inNormal := color.New(color.FgWhite).FprintlnFunc()
	inHeightLight := color.New(color.FgYellow).FprintlnFunc()

	now := time.Now()
	month := now.Month()

	for i, item := range items {
		var (
			strExDate   string
			tExDate     time.Time
			strAmount   string
			strDYield   string
			strWADYield string
		)

		waPrice := item.WeightedAveragePrice()

		if len(item.Stock.Dividends) > 0 {
			d := item.Stock.Dividends[0]

			tExDate = d.ExDate
			strExDate = export.PrintDate(tExDate)

			if item.Stock.Dividends[0].Amount.Amount > 0 {
				wADYield := d.Amount.Amount * 4 / waPrice.Amount * 100

				strAmount = fmt.Sprintf("%.*f", precision, d.Amount.Amount)
				strDYield = fmt.Sprintf("%.*f%%", precision, item.Stock.DividendYield)
				strWADYield = fmt.Sprintf("%.*f%%", precision, wADYield)
			}
		}

		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s\t %s\t %s\t %s\t %s\t",
			i+1,
			item.Stock.Name,
			item.Stock.Exchange.Symbol,
			item.Stock.Symbol,
			item.Amount,
			export.PrintValue(item.Stock.Value, precision),
			export.PrintValue(waPrice, precision),
			strExDate,
			strAmount,
			strDYield,
			strWADYield,
			export.PrintDate(item.Stock.LastPriceUpdate),
		)

		if len(strExDate) > 0 && tExDate.Month() == month {
			inHeightLight(tw, str)
		} else {
			inNormal(tw, str)
		}
	}
}

// formatStockItemsToScreen ...
func formatStockItemsToScreen(tw *tabwriter.Writer, precision int, items []*wallet.Item) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")
	noColor(tw, "# Price")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Price\t WA Price\t High 52wk\t Low 52wk\t Buy Under\t Last Price Update\t")

	normal := color.New(color.FgWhite).FprintlnFunc()
	overSell := color.New(color.FgGreen).FprintlnFunc()
	overBuy := color.New(color.FgRed).FprintlnFunc()

	for i, item := range items {
		waPrice := item.WeightedAveragePrice()

		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s\t %s\t %s\t %s\t",
			i+1,
			item.Stock.Name,
			item.Stock.Exchange.Symbol,
			item.Stock.Symbol,
			item.Amount,
			export.PrintValue(item.Stock.Value, precision),
			export.PrintValue(waPrice, precision),
			export.PrintValue(item.Stock.High52week, precision),
			export.PrintValue(item.Stock.Low52week, precision),
			export.PrintValue(item.Stock.BuyUnder(), precision),
			export.PrintDate(item.Stock.LastPriceUpdate),
		)

		switch item.Stock.ComparePriceWithHighLow() {
		case 1:
			overBuy(tw, str)
		case -1:
			overSell(tw, str)
		default:
			normal(tw, str)
		}
	}
}

// formatWalletItemToScreen - convert an Item structure to screen
func formatWalletItemToScreen(w *wallet.Wallet) *tabwriter.Writer {
	precision := 2
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	sortItems := make([]*wallet.Item, 0, len(w.Items))
	for _, item := range w.Items {
		sortItems = append(sortItems, item)
	}

	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")
	noColor(tw, "# General")
	formatItemsToScreen(tw, precision, w, sortItems)

	noColor(tw, "")
	noColor(tw, "# Dividend")
	formatStockItemsDividendToScreen(tw, precision, sortItems)
	formatStockItemsToScreen(tw, precision, sortItems)
	formatStockItemsOperationToScreen(tw, precision, sortItems)

	fmt.Fprintln(tw, "")

	return tw
}

// formatStockItemsOperationToScreen ...
func formatStockItemsOperationToScreen(tw *tabwriter.Writer, precision int, items []*wallet.Item) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")
	noColor(tw, "# Operation")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Type\t Price\t Exchange\t Commission\t O. Price\t")

	inNormal := color.New(color.FgWhite).FprintlnFunc()

	for _, item := range items {
		for i, o := range item.Operations {
			if o.Action == operation.Buy || o.Action == operation.Sell {
				str := fmt.Sprintf(
					"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s\t %s\t %s\t",
					i+1,
					item.Stock.Name,
					item.Stock.Exchange.Symbol,
					item.Stock.Symbol,
					o.Amount,
					o.Action,
					export.PrintValue(o.Price, precision),
					export.PrintValue(o.PriceChange, 4),
					export.PrintValue(o.FinalCommission(), precision),
					export.PrintValue(o.FinalPricePaid(), precision),
				)

				inNormal(tw, str)
			}
		}
	}
}
