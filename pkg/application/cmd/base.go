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
		DB     *sqlx.DB
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

func (cmd *Base) InitDatabaseConnection() error {
	db, err := sqlx.Connect("postgres", cmd.config.Database.DSN)
	if err != nil {
		return errors.Wrap(err, "Connecting to postgres")
	}

	cmd.DB = db

	return nil
}

func (cmd *Base) Container(db *sqlx.DB) *app.Container {
	return app.NewContainer(cmd.ctx, db, cmd.config, cmd.cache)
}

func (cmd *Base) initCommandBus() *cbus.Bus {
	// Database connection
	if cmd.DB == nil {
		logger.FromContext(cmd.ctx).Fatal("Database connection not initialized")
	}

	// STORAGE
	stockFinder := storage.NewStockFinder(cmd.DB)
	stockDividendFinder := storage.NewStockDividendFinder(cmd.DB)
	walletFinder := storage.NewWalletFinder(cmd.DB)
	marketFinder := storage.NewMarketFinder(cmd.DB)
	exchangeFinder := storage.NewExchangeFinder(cmd.DB)
	stockInfoFinder := storage.NewStockInfoFinder(cmd.DB)
	bankAccountFinder := storage.NewBankAccountFinder(cmd.DB)

	stockPersister := storage.NewStockPersister(cmd.DB)
	walletPersister := storage.NewWalletPersister(cmd.DB)
	stockInfoPersister := storage.NewStockInfoPersister(cmd.DB)
	stockDividendPersister := storage.NewStockDividendPersister(cmd.DB)
	transferPersister := storage.NewTransferPersister(cmd.DB)

	// CLIENT
	//timeout := time.Second * time.Duration(cmd.config.IEXTrading.Timeout)
	//iexClient := iex.NewClient(cmd.newHTTPClient("IEX-TRADING", timeout))

	timeout := time.Second * time.Duration(cmd.config.CurrencyConverter.Timeout)
	ccClient := cc.NewClient(cmd.newHTTPClient("CURRENCY-CONVERTER", timeout), cmd.cache)

	// SERVICE
	//stockPrice := service.NewBasicStockPrice(cmd.ctx, iexClient)
	stockPriceScrapeYahooService := service.NewYahooScrapeStockPrice(cmd.ctx, cmd.config.QuoteScraper.FinanceYahooQuoteURL)
	stockPriceVolatilityMarketChameleonService := service.NewMarketChameleonStockPriceVolatility(cmd.ctx, cmd.config.QuoteScraper.MarketChameleonURL)
	stockDividendMarketChameleonService := service.NewStockDividendMarketChameleon(cmd.ctx, cmd.config.QuoteScraper.MarketChameleonURL)

	// HANDLER
	importStocksHandler := handler.NewImportStock(marketFinder, exchangeFinder, stockInfoFinder, stockInfoPersister)
	updateAllStockPriceHandler := handler.NewUpdateAllStockPrice(stockFinder)
	updateOneStockPrice := handler.NewUpdateOneStockPrice(stockFinder)
	updateAllStockDividendHandler := handler.NewUpdateAllStockDividend(stockFinder)
	updateOneStockDividendHandler := handler.NewUpdateOneStockDividend(stockFinder)
	importTransferHandler := handler.NewImportTransfer(bankAccountFinder)
	importWalletHandler := handler.NewImportWallet(bankAccountFinder)
	importOperationHandler := handler.NewImportOperation(stockFinder)

	// LISTENER
	persisterStock := listener.NewPersisterStock(stockPersister)
	updateStockPrice := listener.NewUpdateStockPrice(stockFinder, stockPriceScrapeYahooService, stockPersister)
	updateStockDividendYield := listener.NewUpdateStockDividendYield(stockDividendFinder, stockPersister)
	updateWalletCapital := listener.NewUpdateWalletCapital(walletFinder, walletPersister, ccClient)
	updateStockPriceVolatility := listener.NewUpdateStockPriceVolatility(stockPriceVolatilityMarketChameleonService, stockPersister)
	updateStockDividend := listener.NewUpdateStockDividend(stockDividendPersister, stockDividendMarketChameleonService)
	persisterTransfer := listener.NewPersisterTransfer(transferPersister, walletFinder, walletPersister)
	persisterWallet := listener.NewPersisterWallet(walletPersister)
	persisterOperation := listener.NewPersisterOperation(walletFinder, stockFinder, walletPersister, ccClient)

	// COMMAND BUS
	bus := cbus.Bus{}

	// USE CASES
	// Import stocks
	importStock := command.ImportStock{}
	bus.Handle(&importStock, importStocksHandler)
	bus.ListenCommand(cbus.Complete, &importStock, persisterStock)
	bus.ListenCommand(cbus.Complete, &importStock, updateStockPrice)
	bus.ListenCommand(cbus.Complete, &importStock, updateStockDividend)
	bus.ListenCommand(cbus.Complete, &importStock, updateStockDividendYield)
	bus.ListenCommand(cbus.Complete, &importStock, updateStockPriceVolatility)

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

	// Update all stock dividends
	updateAllStocksDividend := command.UpdateAllStockDividend{}
	bus.Handle(&updateAllStocksDividend, updateAllStockDividendHandler)
	bus.ListenCommand(cbus.Complete, &updateAllStocksDividend, updateStockDividend)
	bus.ListenCommand(cbus.Complete, &updateAllStocksDividend, updateStockDividendYield)

	// Update one stock dividends
	updateOneStocksDividend := command.UpdateOneStockDividend{}
	bus.Handle(&updateOneStocksDividend, updateOneStockDividendHandler)
	bus.ListenCommand(cbus.Complete, &updateOneStocksDividend, updateStockDividend)
	bus.ListenCommand(cbus.Complete, &updateOneStocksDividend, updateStockDividendYield)

	// import transfer
	importTransfer := command.ImportTransfer{}
	bus.Handle(&importTransfer, importTransferHandler)
	bus.ListenCommand(cbus.Complete, &importTransfer, persisterTransfer)

	// import wallet
	importWallet := command.ImportWallet{}
	bus.Handle(&importWallet, importWalletHandler)
	bus.ListenCommand(cbus.Complete, &importWallet, persisterWallet)

	// import operation
	importOperation := command.ImportOperation{}
	bus.Handle(&importOperation, importOperationHandler)
	bus.ListenCommand(cbus.Complete, &importOperation, persisterOperation)

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
