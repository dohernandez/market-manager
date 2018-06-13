package export_account

import (
	"context"
	"errors"

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

		accountService *account.Service
	}
)

func NewExportWallet(ctx context.Context, sorting export.Sorting, accountService *account.Service) *exportWallet {
	return &exportWallet{
		ctx:            ctx,
		sorting:        sorting,
		accountService: accountService,
	}
}

func (e *exportWallet) Export() error {
	name := e.ctx.Value("wallet").(string)
	if name == "" {
		return errors.New("missing wallet name")
	}

	w, err := e.accountService.FindWalletWithAllActiveItems(name)
	if err != nil {
		return err
	}

	tabw := formatWalletItemsToScreen(w, e.sorting)
	tabw.Flush()

	return nil
}
