package container

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
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
	walletItemFinder    wallet.Finder
	bankAccountFinder   bank.Finder

	stockPersister         stock.Persister
	stockDividendPersister dividend.Persister
	operationPersister     operation.Persister
	walletItemPersister    wallet.Persister
	transferPersister      transfer.Persister

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

func (c *Container) walletItemFinderInstance() wallet.Finder {
	if c.walletItemFinder == nil {
		c.walletItemFinder = storage.NewWalletItemFinder(c.db)
	}

	return c.walletItemFinder
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

func (c *Container) operationPersisterInstance() operation.Persister {
	if c.operationPersister == nil {
		c.operationPersister = storage.NewOperationPersister(c.db)
	}

	return c.operationPersister
}

func (c *Container) walletItemPersisterInstance() wallet.Persister {
	if c.walletItemPersister == nil {
		c.walletItemPersister = storage.NewWalletItemPersister(c.db)
	}

	return c.walletItemPersister
}

func (c *Container) transferPersisterInstance() transfer.Persister {
	if c.transferPersister == nil {
		c.transferPersister = storage.NewTransferPersister(c.db)
	}

	return c.transferPersister
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// SERVICE
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) PurchaseServiceInstance() *purchase.Service {
	if c.purchaseService == nil {
		c.purchaseService = purchase.NewService(
			c.stockPersisterInstance(),
			c.stockFinderInstance(),
			c.stockDividendPersisterInstance(),
			c.stockDividendFinderInstance(),
			c.marketFinderInstance(),
			c.exchangeFinderInstance(),
		)
	}

	return c.purchaseService
}

func (c *Container) AccountServiceInstance() *account.Service {
	if c.accountService == nil {
		c.accountService = account.NewService(
			c.operationPersisterInstance(),
			c.walletItemFinderInstance(),
			c.walletItemPersisterInstance(),
		)
	}

	return c.accountService
}

func (c *Container) BankingServiceInstance() *banking.Service {
	if c.bankingService == nil {
		c.bankingService = banking.NewService(
			c.bankAccountFinderInstance(),
			c.transferPersisterInstance(),
		)
	}

	return c.bankingService
}
