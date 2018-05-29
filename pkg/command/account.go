package command

import (
	"context"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/satori/go.uuid"
	"github.com/urfave/cli"

	"os"

	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

// AccountCommand ...
type AccountCommand struct {
	*BaseCommand
}

// NewAccountCommand constructs AccountCommand
func NewAccountCommand(baseCommand *BaseCommand) *AccountCommand {
	return &AccountCommand{
		BaseCommand: baseCommand,
	}
}

// List in csv format the wallet items from a wallet
func (cmd *AccountCommand) WalletItems(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).WithError(err).Fatal("Missing wallet name")
	}

	w, err := c.AccountServiceInstance().GetWalletWithAllItems(cliCtx.String("wallet"))
	if err != nil {
		return err
	}

	tabw := cmd.formatWalletItemsToScreen(w)
	tabw.Flush()

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// START Wallet Items Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ByName implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type WalletItems []*wallet.Item

func (s WalletItems) Len() int      { return len(s) }
func (s WalletItems) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type WalletItemsByName struct {
	WalletItems
}

func (s WalletItemsByName) Less(i, j int) bool {
	return s.WalletItems[i].Stock.Name < s.WalletItems[j].Stock.Name
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// END Wallet Items Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// formatItemsToScreen - convert Items structure to csv string
func (cmd *AccountCommand) formatWalletItemsToScreen(w *wallet.Wallet) *tabwriter.Writer {
	precision := 2
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	fmt.Fprintln(tw, "")
	cmd.formatWalletToScreen(tw, precision, w)

	fmt.Fprintln(tw, "")

	cmd.formatItemsToScreen(tw, precision, w.Items)
	fmt.Fprintln(tw, "")

	return tw
}

// formatItemsToScreen - convert wallet structure to csv string
func (cmd *AccountCommand) formatWalletToScreen(tw *tabwriter.Writer, precision int, w *wallet.Wallet) {
	fmt.Fprintln(tw, "Invested\t Capital\t Funds\t Benefits\t % Benefits\t")
	str := fmt.Sprintf(
		"%.*f\t %.*f\t %.*f\t %.*f\t %.*f%%\t",
		precision,
		w.Invested.Amount,
		precision,
		w.Capital.Amount,
		precision,
		w.Funds.Amount,
		precision,
		w.NetBenefits().Amount,
		precision,
		w.PercentageBenefits(),
	)
	fmt.Fprintln(tw, str)
}

// formatItemsToScreen - convert Items structure to csv string
func (cmd *AccountCommand) formatItemsToScreen(tw *tabwriter.Writer, precision int, items map[uuid.UUID]*wallet.Item) {
	sortItems := make([]*wallet.Item, 0, len(items))

	for _, item := range items {
		sortItems = append(sortItems, item)
	}

	sort.Sort(WalletItemsByName{sortItems})

	fmt.Fprintln(tw, "#\t Stock\t Market\t Symbol\t Amount\t Capital\t Invested\t Dividend\t Buys\t Sells\t Benefits\t % Benefits\t")
	for i, item := range sortItems {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %d\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f\t %.*f%%\t",
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
			item.Sells.Amount,
			precision,
			item.NetBenefits().Amount,
			precision,
			item.PercentageBenefits(),
		)
		fmt.Fprintln(tw, str)
	}
}
