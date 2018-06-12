package export_account

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/satori/go.uuid"

	"github.com/fatih/color"

	"github.com/dohernandez/market-manager/pkg/export"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

// formatWalletItemsToScreen - convert Items structure to screen
func formatWalletItemsToScreen(w *wallet.Wallet, sorting export.Sorting) *tabwriter.Writer {
	precision := 2
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	formatWalletToScreen(tw, precision, w)

	formatItemsToScreen(tw, precision, w.Items, sorting)

	fmt.Fprintln(tw, "")

	return tw
}

// formatWalletToScreen - convert wallet structure to screen
func formatWalletToScreen(tw *tabwriter.Writer, precision int, w *wallet.Wallet) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "Invested\t Capital\t Funds\t Net Capital\t Net Benefits\t % Benefits\t Dividends\t Connection\t Interest\t Commisions\t")

	pColor := color.New(color.FgGreen).FprintlnFunc()
	if w.PercentageBenefits() < 0 {
		pColor = color.New(color.FgRed).FprintlnFunc()
	}

	str := fmt.Sprintf(
		"%s\t %s\t %s\t %s\t %s\t %.*f%%\t %s\t %s\t %s\t %s\t",
		export.PrintValue(w.Invested, precision),
		export.PrintValue(w.Capital, precision),
		export.PrintValue(w.Funds, precision),
		export.PrintValue(w.NetCapital(), precision),
		export.PrintValue(w.NetBenefits(), precision),
		precision,
		w.PercentageBenefits(),
		export.PrintValue(w.Dividend, precision),
		export.PrintValue(w.Connection, precision),
		export.PrintValue(w.Interest, precision),
		export.PrintValue(w.Commission, precision),
	)
	pColor(tw, str)
}

// formatItemsToScreen - convert Items structure to screen
func formatItemsToScreen(tw *tabwriter.Writer, precision int, items map[uuid.UUID]*wallet.Item, sorting export.Sorting) {
	sortItems := make([]*wallet.Item, 0, len(items))

	for _, item := range items {
		sortItems = append(sortItems, item)
	}

	switch sorting.By {
	case Invested:
		sort.Sort(WalletItemsByInvested{sortItems})
	default:
		sort.Sort(WalletItemsByName{sortItems})
	}

	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Capital\t Invested\t Dividend\t Buys\t Sells\t Benefits\t % Benefits\t Change\t")

	inProfits := color.New(color.FgGreen).FprintlnFunc()
	inLooses := color.New(color.FgRed).FprintlnFunc()

	for i, item := range sortItems {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %s\t %s\t %s\t %s\t %s\t %s\t %.*f%%\t %s\t",
			i+1,
			item.Stock.Name,
			item.Stock.Exchange.Symbol,
			item.Stock.Symbol,
			item.Amount,
			export.PrintValue(item.Capital(), precision),
			export.PrintValue(item.Invested, precision),
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
