package container

import (
	"context"

	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock/dividend"
	"github.com/dohernandez/market-manager/pkg/storage"
	"github.com/jmoiron/sqlx"
)

type Container struct {
	ctx    context.Context
	db     *sqlx.DB
	config *config.Specification

	marketFinder        market.Finder
	exchangeFinder      exchange.Finder
	stockFinder         stock.Finder
	stockDividendFinder dividend.Finder

	stockPersister         stock.Persister
	stockDividendPersister dividend.Persister
	accountPersister       account.Persister

	stockService   *stock.Service
	accountService *account.Service
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

func (c *Container) MarketFinderInstance() market.Finder {
	if c.marketFinder == nil {
		c.marketFinder = storage.NewMarketFinder(c.db)
	}

	return c.marketFinder
}

func (c *Container) ExchangeFinderInstance() exchange.Finder {
	if c.exchangeFinder == nil {
		c.exchangeFinder = storage.NewExchangeFinder(c.db)
	}

	return c.exchangeFinder
}

func (c *Container) StockFinderInstance() stock.Finder {
	if c.stockFinder == nil {
		c.stockFinder = storage.NewStockFinder(c.db)
	}

	return c.stockFinder
}

func (c *Container) StockDividendFinderInstance() dividend.Finder {
	if c.stockDividendFinder == nil {
		c.stockDividendFinder = storage.NewStockDividendFinder(c.db)
	}

	return c.stockDividendFinder
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// PERSISTER
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) StockPersisterInstance() stock.Persister {
	if c.stockPersister == nil {
		c.stockPersister = storage.NewStockPersister(c.db)
	}

	return c.stockPersister
}

func (c *Container) StockDividendPersisterInstance() dividend.Persister {
	if c.stockDividendPersister == nil {
		c.stockDividendPersister = storage.NewStockDividendPersister(c.db)
	}

	return c.stockDividendPersister
}

func (c *Container) AccountPersisterInstance() account.Persister {
	if c.accountPersister == nil {
		c.accountPersister = storage.NewAccountPersister(c.db)
	}

	return c.accountPersister
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// SERVICE
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func (c *Container) StockServiceInstance() *stock.Service {
	if c.stockService == nil {
		c.stockService = stock.NewService(
			c.StockPersisterInstance(),
			c.StockFinderInstance(),
			c.StockDividendPersisterInstance(),
			c.StockDividendFinderInstance(),
		)
	}

	return c.stockService
}

func (c *Container) AccountServiceInstance() *account.Service {
	if c.accountService == nil {
		c.accountService = account.NewService(
			c.AccountPersisterInstance(),
		)
	}

	return c.accountService
}
