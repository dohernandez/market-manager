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
	// BaseCMD hold common command properties
	BaseCMD struct {
		ctx    context.Context
		config *config.Specification
		cache  *cache.Cache
	}
)

// NewBaseCMD creates a structure with common shared properties of the commands
func NewBaseCMD(ctx context.Context, config *config.Specification) *BaseCMD {
	return &BaseCMD{
		ctx:    ctx,
		config: config,
		cache:  cache.New(time.Hour*2, time.Hour*10),
	}
}

func (cmd *BaseCMD) initDatabaseConnection() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cmd.config.Database.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "Connecting to postgres")
	}

	return db, nil
}

func (cmd *BaseCMD) Container(db *sqlx.DB) *app.Container {
	return app.NewContainer(cmd.ctx, db, cmd.config, cmd.cache)
}

func (cmd *BaseCMD) initCommandBus() *cbus.Bus {
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
	//stockPrice := service.NewStockPrice(cmd.ctx, iexClient)
	stockPriceScrapeYahoo := service.NewStockPriceScrapeYahoo(cmd.ctx, cmd.config.QuoteScraper.FinanceYahooQuoteURL)

	// HANDLER
	updateAllStockPriceHandler := handler.NewUpdateAllStockPrice(stockFinder, stockPriceScrapeYahoo, stockPersister)

	// LISTENER
	updateStockDividendYield := listener.NewUpdateStockDividendYield(stockDividendFinder, stockPersister)
	updateWalletCapital := listener.NewUpdateWalletCapital(walletFinder, walletPersister, ccClient)

	// COMMAND BUS
	bus := cbus.Bus{}

	// Update all stock price
	updateAllStocksPrice := command.UpdateAllStocksPrice{}
	bus.Handle(&updateAllStocksPrice, updateAllStockPriceHandler)
	bus.ListenCommand(cbus.Complete, &updateAllStocksPrice, updateStockDividendYield)
	bus.ListenCommand(cbus.Complete, &updateAllStocksPrice, updateWalletCapital)

	// Update all stock price
	bus.Handle(&command.UpdateOneStockPrice{}, handler.NewUpdateOneStockPrice(stockFinder))

	return &bus
}

func (cmd *BaseCMD) newHTTPClient(name string, timeout time.Duration) *http.Client {
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
