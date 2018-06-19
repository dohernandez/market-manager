package export_account

import (
	"context"
	"errors"

	"text/tabwriter"

	"github.com/dohernandez/market-manager/pkg/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/export"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

const (
	Stock    export.SortBy = "stock"
	Invested export.SortBy = "invested"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// START Wallet Items Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type WalletItems []*wallet.Item

func (s WalletItems) Len() int      { return len(s) }
func (s WalletItems) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// WalletItemsByName implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded ExportWalletItems value.
type WalletItemsByName struct {
	WalletItems
}

func (s WalletItemsByName) Less(i, j int) bool {
	return s.WalletItems[i].Stock.Name < s.WalletItems[j].Stock.Name
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
// END Wallet Items Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	exportWallet struct {
		ctx     context.Context
		sorting export.Sorting
		config  *config.Specification

		accountService *account.Service

		ccClient *cc.Client
	}
)

func NewExportWallet(ctx context.Context, sorting export.Sorting, config *config.Specification, accountService *account.Service, ccClient *cc.Client) *exportWallet {
	return &exportWallet{
		ctx:            ctx,
		sorting:        sorting,
		accountService: accountService,
		ccClient:       ccClient,
		config:         config,
	}
}

func (e *exportWallet) Export() error {
	name := e.ctx.Value("wallet").(string)
	if name == "" {
		return errors.New("missing wallet name")
	}

	cEURUSD, err := e.ccClient.Converter.Get()
	if err != nil {
		return err
	}

	w, err := e.accountService.FindWalletByName(name)
	if err != nil {
		return err
	}
	w.SetCapitalRate(cEURUSD.EURUSD)

	tabw := new(tabwriter.Writer)
	stkSymbol := e.ctx.Value("stock").(string)

	if stkSymbol == "" {
		err := e.accountService.LoadActiveWalletItems(w)
		if err != nil {
			return err
		}

		tabw = formatWalletItemsToScreen(w, e.sorting, e.config.Retention)
	} else {
		err := e.accountService.LoadWalletItem(w, stkSymbol)
		if err != nil {
			return err
		}

		tabw = formatWalletItemToScreen(w)
	}

	tabw.Flush()

	return nil
}
