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

	fmt.Fprintln(tw, "")
	formatWalletToScreen(tw, precision, w)

	formatItemsToScreen(tw, precision, w.Items, sorting)
	fmt.Fprintln(tw, "")

	return tw
}

// formatWalletToScreen - convert wallet structure to screen
func formatWalletToScreen(tw *tabwriter.Writer, precision int, w *wallet.Wallet) {
	fmt.Fprintln(tw, "Invested\t Capital\t Funds\t Net Capital\t Net Benefits\t % Benefits\t Dividends\t Connection\t Interest\t Commisions\t")
	str := fmt.Sprintf(
		"%.*f\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f%%\t %.*f\t %.*f\t %.*f\t %.*f\t",
		precision,
		w.Invested.Amount,
		precision,
		w.Capital.Amount,
		precision,
		w.Funds.Amount,
		precision,
		w.NetCapital().Amount,
		precision,
		w.NetBenefits().Amount,
		precision,
		w.PercentageBenefits(),
		precision,
		w.Dividend.Amount,
		precision,
		w.Connection.Amount,
		precision,
		w.Interest.Amount,
		precision,
		w.Commission.Amount,
	)
	fmt.Fprintln(tw, str)
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

	header := color.New(color.Bold, color.FgBlack).FprintlnFunc()
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Capital\t Invested\t Dividend\t Buys\t WA Price\t Sells\t Benefits\t % Benefits\t Change\t")

	inProfits := color.New(color.Bold, color.FgGreen).FprintlnFunc()
	inLooses := color.New(color.Bold, color.FgRed).FprintlnFunc()

	for i, item := range sortItems {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f%%\t %.*f\t",
			i+1,
			item.Stock.Name,
			item.Stock.Exchange.Symbol,
			item.Stock.Symbol,
			item.Amount,
			precision,
			item.Capital().Amount,
			precision,
			item.Invested.Amount,
			precision,
			item.Dividend.Amount,
			precision,
			item.Buys.Amount,
			precision,
			item.WeightedAveragePrice().Amount,
			precision,
			item.Sells.Amount,
			precision,
			item.NetBenefits().Amount,
			precision,
			item.PercentageBenefits(),
			precision,
			item.Change().Amount,
		)

		if item.PercentageBenefits() > 0 {
			inProfits(tw, str)
		} else {
			inLooses(tw, str)
		}
	}
}
