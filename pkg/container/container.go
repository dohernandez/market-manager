package container

import (
	"context"

	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/market-manager/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
	"github.com/dohernandez/market-manager/pkg/storage"
	"github.com/jmoiron/sqlx"
)

type Container struct {
	ctx    context.Context
	db     *sqlx.DB
	config *config.Specification

	marketFinder   market.Finder
	exchangeFinder exchange.Finder

	stockPersister stock.Persister
}

func NewContainer(ctx context.Context, db *sqlx.DB, config *config.Specification) *Container {
	return &Container{
		ctx:    ctx,
		db:     db,
		config: config,
	}
}

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

func (c *Container) StockPersisterInstance() stock.Persister {
	if c.stockPersister == nil {
		c.stockPersister = storage.NewStockPersister(c.db)
	}

	return c.stockPersister
}
