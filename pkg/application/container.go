package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/f2prateek/train"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"

	"github.com/dohernandez/market-manager/pkg/application/config"
	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/application/storage"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
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
	stockInfoPersister     stock.InfoPersister
	stockInfoFinder        stock.InfoFinder

	iexClient *iex.Client
	ccClient  *cc.Client

	purchaseService *service.Purchase
	accountService  *service.Account
	bankingService  *service.Banking

	cache *cache.Cache
}

func NewContainer(ctx context.Context, db *sqlx.DB, config *config.Specification, ch *cache.Cache) *Container {
	return &Container{
		ctx:    ctx,
		db:     db,
		config: config,
		cache:  ch,
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

func (c *Container) stockInfoFinderInstance() stock.InfoFinder {
	if c.stockInfoFinder == nil {
		c.stockInfoFinder = storage.NewStockInfoFinder(c.db)
	}

	return c.stockInfoFinder
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

func (c *Container) stockInfoPersisterInstance() stock.InfoPersister {
	if c.stockInfoPersister == nil {
		c.stockInfoPersister = storage.NewStockInfoPersister(c.db)
	}

	return c.stockInfoPersister
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

		c.ccClient = cc.NewClient(c.newHTTPClient("CURRENCY-CONVERTER", timeout), c.cache)
	}

	return c.ccClient
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// SERVICE
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) PurchaseServiceInstance() *service.Purchase {
	if c.purchaseService == nil {
		c.purchaseService = service.NewPurchaseService(
			c.ctx,
			c.stockPersisterInstance(),
			c.stockFinderInstance(),
			c.stockDividendPersisterInstance(),
			c.stockDividendFinderInstance(),
			c.marketFinderInstance(),
			c.exchangeFinderInstance(),
			c.AccountServiceInstance(),
			c.IEXTradingClientInstance(),
			c.stockInfoFinderInstance(),
			c.stockInfoPersisterInstance(),
		)
	}

	return c.purchaseService
}

func (c *Container) AccountServiceInstance() *service.Account {
	if c.accountService == nil {
		c.accountService = service.NewAccountService(
			c.walletFinderInstance(),
			c.walletPersisterInstance(),
			c.stockFinderInstance(),
			c.CurrencyConverterClientInstance(),
			c.stockDividendFinderInstance(),
		)
	}

	return c.accountService
}

func (c *Container) BankingServiceInstance() *service.Banking {
	if c.bankingService == nil {
		c.bankingService = service.NewBankingService(
			c.bankAccountFinderInstance(),
			c.transferPersisterInstance(),
			c.AccountServiceInstance(),
		)
	}

	return c.bankingService
}
