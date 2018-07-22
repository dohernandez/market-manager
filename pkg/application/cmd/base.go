package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/f2prateek/train"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/application"
	"github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/config"
	"github.com/dohernandez/market-manager/pkg/application/handler"
	"github.com/dohernandez/market-manager/pkg/application/listener"
	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/application/storage"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
)

type (
	// Base hold common command properties
	Base struct {
		ctx    context.Context
		config *config.Specification
		cache  *cache.Cache
	}
)

// NewBase creates a structure with common shared properties of the commands
func NewBase(ctx context.Context, config *config.Specification) *Base {
	return &Base{
		ctx:    ctx,
		config: config,
		cache:  cache.New(time.Hour*2, time.Hour*10),
	}
}

func (cmd *Base) initDatabaseConnection() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cmd.config.Database.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "Connecting to postgres")
	}

	return db, nil
}

func (cmd *Base) Container(db *sqlx.DB) *app.Container {
	return app.NewContainer(cmd.ctx, db, cmd.config, cmd.cache)
}

func (cmd *Base) initCommandBus() *cbus.Bus {
	// Database connection
	logger.FromContext(cmd.ctx).Info("Initializing database connection")

	db, err := sqlx.Connect("postgres", cmd.config.Database.DSN)
	if err != nil {
		logger.FromContext(cmd.ctx).WithError(err).Fatal("Failed initializing database")
	}

	// STORAGE
	stockFinder := storage.NewStockFinder(db)
	stockDividendFinder := storage.NewStockDividendFinder(db)
	walletFinder := storage.NewWalletFinder(db)

	stockPersister := storage.NewStockPersister(db)
	walletPersister := storage.NewWalletPersister(db)

	// CLIENT
	//timeout := time.Second * time.Duration(cmd.config.IEXTrading.Timeout)
	//iexClient := iex.NewClient(cmd.newHTTPClient("IEX-TRADING", timeout))

	timeout := time.Second * time.Duration(cmd.config.CurrencyConverter.Timeout)
	ccClient := cc.NewClient(cmd.newHTTPClient("CURRENCY-CONVERTER", timeout), cmd.cache)

	// SERVICE
	//stockPrice := service.NewBasicStockPrice(cmd.ctx, iexClient)
	stockPriceScrapeYahoo := service.NewYahooScrapeStockPrice(cmd.ctx, cmd.config.QuoteScraper.FinanceYahooQuoteURL)
	stockPriceVolatilityMarketChameleon := service.NewMarketChameleonStockPriceVolatility(cmd.ctx, cmd.config.QuoteScraper.MarketChameleonURL)

	// HANDLER
	updateAllStockPriceHandler := handler.NewUpdateAllStockPrice(stockFinder)
	updateOneStockPrice := handler.NewUpdateOneStockPrice(stockFinder)

	// LISTENER

	updateStockPrice := listener.NewUpdateStockPrice(stockFinder, stockPriceScrapeYahoo, stockPersister)
	updateStockDividendYield := listener.NewUpdateStockDividendYield(stockDividendFinder, stockPersister)
	updateWalletCapital := listener.NewUpdateWalletCapital(walletFinder, walletPersister, ccClient)
	updateStockPriceVolatility := listener.NewUpdateStockPriceVolatility(stockPriceVolatilityMarketChameleon, stockPersister)

	// COMMAND BUS
	bus := cbus.Bus{}

	// Update all stock price
	updateAllStocksPrice := command.UpdateAllStocksPrice{}
	bus.Handle(&updateAllStocksPrice, updateAllStockPriceHandler)
	bus.ListenCommand(cbus.Complete, &updateAllStocksPrice, updateStockPrice)
	bus.ListenCommand(cbus.Complete, &updateAllStocksPrice, updateStockDividendYield)
	bus.ListenCommand(cbus.Complete, &updateAllStocksPrice, updateWalletCapital)
	bus.ListenCommand(cbus.Complete, &updateAllStocksPrice, updateStockPriceVolatility)

	// Update one stock price
	updateOneStocksPrice := command.UpdateOneStockPrice{}
	bus.Handle(&updateOneStocksPrice, updateOneStockPrice)
	bus.ListenCommand(cbus.Complete, &updateOneStocksPrice, updateStockPrice)
	bus.ListenCommand(cbus.Complete, &updateOneStocksPrice, updateStockDividendYield)
	bus.ListenCommand(cbus.Complete, &updateOneStocksPrice, updateWalletCapital)
	bus.ListenCommand(cbus.Complete, &updateOneStocksPrice, updateStockPriceVolatility)

	return &bus
}

func (cmd *Base) newHTTPClient(name string, timeout time.Duration) *http.Client {
	clt := http.Client{}

	// Add middleware
	st := gobreaker.Settings{
		Name: fmt.Sprintf("%s Client Circuit breaker", name),
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.FromContext(cmd.ctx).WithFields(logrus.Fields{
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
