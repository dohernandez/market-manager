package container

import (
	"context"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"

	"github.com/f2prateek/train"

	"github.com/dohernandez/market-manager/pkg/client"
	"github.com/dohernandez/market-manager/pkg/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
	"github.com/dohernandez/market-manager/pkg/storage"
)

type Container struct {
	ctx    context.Context
	db     *sqlx.DB
	config *config.Specification

	marketFinder        market.Finder
	exchangeFinder      exchange.Finder
	stockFinder         stock.Finder
	stockDividendFinder dividend.Finder
	walletFinder        wallet.Finder
	bankAccountFinder   bank.Finder

	stockPersister         stock.Persister
	stockDividendPersister dividend.Persister
	walletPersister        wallet.Persister
	transferPersister      transfer.Persister

	importStorage _import.Storage

	iexClient *iex.Client
	ccClient  *cc.Client

	purchaseService *purchase.Service
	accountService  *account.Service
	bankingService  *banking.Service
}

func NewContainer(ctx context.Context, db *sqlx.DB, config *config.Specification) *Container {
	return &Container{
		ctx:    ctx,
		db:     db,
		config: config,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// FINDER
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *Container) marketFinderInstance() market.Finder {
	if c.marketFinder == nil {
		c.marketFinder = storage.NewMarketFinder(c.db)
	}

	return c.marketFinder
}

func (c *Container) exchangeFinderInstance() exchange.Finder {
	if c.exchangeFinder == nil {
		c.exchangeFinder = storage.NewExchangeFinder(c.db)
	}

	return c.exchangeFinder
}

func (c *Container) stockFinderInstance() stock.Finder {
	if c.stockFinder == nil {
		c.stockFinder = storage.NewStockFinder(c.db)
	}

	return c.stockFinder
}

func (c *Container) stockDividendFinderInstance() dividend.Finder {
	if c.stockDividendFinder == nil {
		c.stockDividendFinder = storage.NewStockDividendFinder(c.db)
	}

	return c.stockDividendFinder
}

func (c *Container) walletFinderInstance() wallet.Finder {
	if c.walletFinder == nil {
		c.walletFinder = storage.NewWalletFinder(c.db)
	}

	return c.walletFinder
}

func (c *Container) bankAccountFinderInstance() bank.Finder {
	if c.bankAccountFinder == nil {
		c.bankAccountFinder = storage.NewBankAccountFinder(c.db)
	}

	return c.bankAccountFinder
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// PERSISTER
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) stockPersisterInstance() stock.Persister {
	if c.stockPersister == nil {
		c.stockPersister = storage.NewStockPersister(c.db)
	}

	return c.stockPersister
}

func (c *Container) stockDividendPersisterInstance() dividend.Persister {
	if c.stockDividendPersister == nil {
		c.stockDividendPersister = storage.NewStockDividendPersister(c.db)
	}

	return c.stockDividendPersister
}

func (c *Container) walletPersisterInstance() wallet.Persister {
	if c.walletPersister == nil {
		c.walletPersister = storage.NewWalletPersister(c.db)
	}

	return c.walletPersister
}

func (c *Container) transferPersisterInstance() transfer.Persister {
	if c.transferPersister == nil {
		c.transferPersister = storage.NewTransferPersister(c.db)
	}

	return c.transferPersister
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// STORAGE
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) ImportStorageInstance() _import.Storage {
	if c.importStorage == nil {
		c.importStorage = storage.NewImportStorage(c.db)
	}

	return c.importStorage
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// CLIENT
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) newHTTPClient(name string, timeout time.Duration) *http.Client {
	clt := http.Client{}

	// Add middleware
	st := gobreaker.Settings{
		Name: fmt.Sprintf("%s Client Circuit breaker", name),
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.FromContext(c.ctx).WithFields(logrus.Fields{
				"name": name,
				"from": from,
				"to":   to,
			}).Error("Circuit breaker state changed")
		},
	}
	cbInterceptor := client.NewCircuitBreaker(st)

	clt.Timeout = timeout

	if clt.Transport != nil {
		clt.Transport = train.TransportWith(clt.Transport, cbInterceptor)
		return &clt
	}

	clt.Transport = train.Transport(cbInterceptor)

	return &clt
}

func (c *Container) IEXTradingClientInstance() *iex.Client {
	if c.iexClient == nil {
		timeout := time.Second * time.Duration(c.config.IEXTrading.Timeout)

		c.iexClient = iex.NewClient(c.newHTTPClient("IEX-TRADING", timeout))
	}

	return c.iexClient
}

func (c *Container) CurrencyConverterClientInstance() *cc.Client {
	if c.ccClient == nil {
		timeout := time.Second * time.Duration(c.config.CurrencyConverter.Timeout)

		c.ccClient = cc.NewClient(c.newHTTPClient("CURRENCY-CONVERTER", timeout))
	}

	return c.ccClient
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// SERVICE
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) PurchaseServiceInstance() *purchase.Service {
	if c.purchaseService == nil {
		c.purchaseService = purchase.NewService(
			c.ctx,
			c.stockPersisterInstance(),
			c.stockFinderInstance(),
			c.stockDividendPersisterInstance(),
			c.stockDividendFinderInstance(),
			c.marketFinderInstance(),
			c.exchangeFinderInstance(),
			c.AccountServiceInstance(),
			c.IEXTradingClientInstance(),
		)
	}

	return c.purchaseService
}

func (c *Container) AccountServiceInstance() *account.Service {
	if c.accountService == nil {
		c.accountService = account.NewService(
			c.walletFinderInstance(),
			c.walletPersisterInstance(),
			c.stockFinderInstance(),
			c.CurrencyConverterClientInstance(),
		)
	}

	return c.accountService
}

func (c *Container) BankingServiceInstance() *banking.Service {
	if c.bankingService == nil {
		c.bankingService = banking.NewService(
			c.bankAccountFinderInstance(),
			c.transferPersisterInstance(),
			c.AccountServiceInstance(),
		)
	}

	return c.bankingService
}
