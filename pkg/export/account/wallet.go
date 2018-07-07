package export_account

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/application/config"
	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/export"
	"github.com/dohernandez/market-manager/pkg/market-manager"
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

		accountService *service.Account

		ccClient *cc.Client
	}
)

func NewExportWallet(ctx context.Context, sorting export.Sorting, config *config.Specification, accountService *service.Account, ccClient *cc.Client) *exportWallet {
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

	var retention float64
	if w.Name == "degiro" {
		retention = e.config.Degiro.Retention
	}

	stkSymbol := e.ctx.Value("stock").(string)
	tabw := new(tabwriter.Writer)

	if stkSymbol == "" {
		sells := map[string]int{}
		strSells := e.ctx.Value("sells").(string)
		if strSells != "" {
			sSells := strings.Split(strSells, ",")

			for _, sSell := range sSells {
				sa := strings.Split(sSell, ":")
				a, _ := strconv.Atoi(sa[1])
				sells[sa[0]] = a
			}
		}

		buys := map[string]int{}
		strBuys := e.ctx.Value("buys").(string)
		if strBuys != "" {
			sBuys := strings.Split(strBuys, ",")

			for _, sBuy := range sBuys {
				ba := strings.Split(sBuy, ":")
				a, _ := strconv.Atoi(ba[1])
				buys[ba[0]] = a
			}
		}

		changeCommissions := e.getChangeCommissionsToApplyStockOperation()
		appCommissions, err := e.getAppCommissionsToApplyStockOperation()
		if err != nil {
			return err
		}

		err = e.accountService.LoadActiveWalletItems(w)
		if err != nil {
			return err
		}

		e.accountService.SellStocksWallet(w, sells, changeCommissions, appCommissions)
		e.accountService.BuyStocksWallet(w, buys, changeCommissions, appCommissions)

		items := map[uuid.UUID]*wallet.Item{}
		for k, item := range w.Items {
			if item.Amount == 0 {
				continue
			}

			items[k] = item
		}

		w.Items = items

		tabw = formatWalletItemsToScreen(w, e.sorting, retention)
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

func (e *exportWallet) getChangeCommissionsToApplyStockOperation() map[string]mm.Value {
	pChangeCommissions := map[string]mm.Value{}

	exchanges := e.config.Degiro.Exchanges

	pChangeCommissions["NASDAQ"] = mm.Value{
		Amount:   exchanges.NASDAQ.ChangeCommission.Amount,
		Currency: mm.Currency(exchanges.NASDAQ.ChangeCommission.Currency),
	}

	pChangeCommissions["NYSE"] = mm.Value{
		Amount:   exchanges.NYSE.ChangeCommission.Amount,
		Currency: mm.Currency(exchanges.NYSE.ChangeCommission.Currency),
	}

	return pChangeCommissions
}

func (e *exportWallet) getAppCommissionsToApplyStockOperation() (map[string]service.AppCommissions, error) {
	appCommissions := map[string]service.AppCommissions{}

	exchanges := e.config.Degiro.Exchanges

	// NASDAQ
	cCommission, err := json.Marshal(exchanges.NASDAQ)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not marshal config commission")
	}

	var appCommission service.AppCommissions
	err = json.Unmarshal(cCommission, &appCommission)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not unmarshal config commission")
	}

	appCommissions["NASDAQ"] = appCommission

	// NYSE
	cCommission, err = json.Marshal(exchanges.NYSE)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not marshal config commission")
	}

	appCommission = service.AppCommissions{}
	err = json.Unmarshal(cCommission, &appCommission)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not unmarshal config commission")
	}

	appCommissions["NASDAQ"] = appCommission

	return appCommissions, nil
}
