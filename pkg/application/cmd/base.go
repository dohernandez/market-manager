package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/f2prateek/train"
	"github.com/gogolfing/cbus"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"

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
	walletReload := storage.NewWalletReload(cmd.DB)
	resourceStorage := storage.NewUtilImportStorage(cmd.DB)

	stockPersister := storage.NewStockPersister(cmd.DB)
	walletPersister := storage.NewWalletPersister(cmd.DB)
	stockInfoPersister := storage.NewStockInfoPersister(cmd.DB)
	stockDividendPersister := storage.NewStockDividendPersister(cmd.DB)
	transferPersister := storage.NewTransferPersister(cmd.DB)

	// CLIENT
	timeout := time.Second * time.Duration(cmd.config.CurrencyConverter.Timeout)
	ccClient := cc.NewClient(cmd.newHTTPClient("CURRENCY-CONVERTER", timeout), cmd.cache)

	// SERVICE
	//stockPrice := service.NewBasicStockPrice(cmd.ctx, iexClient)
	stockPriceScrapeYahooService := service.NewYahooScrapeStockPrice(cmd.ctx, cmd.config.QuoteScraper.FinanceYahooQuoteURL)
	stockPriceVolatilityMarketChameleonService := service.NewMarketChameleonStockPriceVolatility(cmd.ctx, cmd.config.QuoteScraper.MarketChameleonURL)
	stockDividendMarketChameleonService := service.NewStockDividendMarketChameleon(cmd.ctx, cmd.config.QuoteScraper.MarketChameleonURL)

	// HANDLER
	importStocksHandler := handler.NewImportStock(marketFinder, exchangeFinder, stockInfoFinder, stockPersister, stockInfoPersister)
	updateAllStockPriceHandler := handler.NewUpdateAllStockPrice(stockFinder)
	updateOneStockPrice := handler.NewUpdateOneStockPrice(stockFinder)
	updateAllStockDividendHandler := handler.NewUpdateAllStockDividend(stockFinder)
	updateOneStockDividendHandler := handler.NewUpdateOneStockDividend(stockFinder)
	importTransferHandler := handler.NewImportTransfer(bankAccountFinder, transferPersister, walletFinder, walletPersister)
	importWalletHandler := handler.NewImportWallet(bankAccountFinder, walletPersister)
	importOperationHandler := handler.NewImportOperation(stockFinder)
	listStockHandler := handler.NewListStock(stockFinder, stockDividendFinder)
	walletDetailsHandler := handler.NewWalletDetails(walletFinder, stockFinder, stockDividendFinder, ccClient, cmd.config.Degiro.Retention)
	updateWalletStocksPriceHandler := handler.NewUpdateWalletStocksPrice(walletFinder, stockFinder)
	reloadWalletHandler := handler.NewReloadWallet(walletFinder, walletReload)
	importRetentionHandler := handler.NewImportRetention(stockFinder, walletFinder, walletPersister)
	addOperationHandler := handler.NewAddOperation(stockFinder)
	walletDateDetailsHandler := handler.NewWalletDateDetails(walletFinder, stockFinder, stockDividendFinder, ccClient, cmd.config.Degiro.Retention, bankAccountFinder)

	// LISTENER
	updateStockPrice := listener.NewUpdateStockPrice(stockFinder, stockPriceScrapeYahooService, stockPersister)
	updateStockDividendYield := listener.NewUpdateStockDividendYield(stockDividendFinder, stockPersister)
	updateWalletCapital := listener.NewUpdateWalletCapital(walletFinder, walletPersister, ccClient)
	updateStockPriceVolatility := listener.NewUpdateStockPriceVolatility(stockPriceVolatilityMarketChameleonService, stockPersister)
	updateStockDividend := listener.NewUpdateStockDividend(stockDividendPersister, stockDividendMarketChameleonService)
	addWalletOperation := listener.NewAddWalletOperation(stockFinder, walletFinder, walletPersister, ccClient)
	registerWalletOperationImport := listener.NewRegisterWalletOperationImport(resourceStorage, cmd.config.Import.AccountsPath)

	// COMMAND BUS
	bus := cbus.Bus{}

	// USE CASES
	// Import stocks
	importStock := command.ImportStock{}
	bus.Handle(&importStock, importStocksHandler)
	bus.ListenCommand(cbus.AfterSuccess, &importStock, updateStockPrice)
	bus.ListenCommand(cbus.AfterSuccess, &importStock, updateStockDividend)
	bus.ListenCommand(cbus.AfterSuccess, &importStock, updateStockDividendYield)
	bus.ListenCommand(cbus.AfterSuccess, &importStock, updateStockPriceVolatility)

	// Update all stock price
	updateAllStocksPrice := command.UpdateAllStocksPrice{}
	bus.Handle(&updateAllStocksPrice, updateAllStockPriceHandler)
	bus.ListenCommand(cbus.AfterSuccess, &updateAllStocksPrice, updateStockPrice)
	bus.ListenCommand(cbus.AfterSuccess, &updateAllStocksPrice, updateStockDividendYield)
	bus.ListenCommand(cbus.AfterSuccess, &updateAllStocksPrice, updateWalletCapital)
	//bus.ListenCommand(cbus.AfterSuccess, &updateAllStocksPrice, updateStockPriceVolatility)

	// Update one stock price
	updateOneStocksPrice := command.UpdateOneStockPrice{}
	bus.Handle(&updateOneStocksPrice, updateOneStockPrice)
	bus.ListenCommand(cbus.AfterSuccess, &updateOneStocksPrice, updateStockPrice)
	bus.ListenCommand(cbus.AfterSuccess, &updateOneStocksPrice, updateStockDividendYield)
	bus.ListenCommand(cbus.AfterSuccess, &updateOneStocksPrice, updateWalletCapital)
	bus.ListenCommand(cbus.AfterSuccess, &updateOneStocksPrice, updateStockPriceVolatility)

	// Update all stock dividends
	updateAllStocksDividend := command.UpdateAllStockDividend{}
	bus.Handle(&updateAllStocksDividend, updateAllStockDividendHandler)
	bus.ListenCommand(cbus.AfterSuccess, &updateAllStocksDividend, updateStockDividend)
	bus.ListenCommand(cbus.AfterSuccess, &updateAllStocksDividend, updateStockDividendYield)

	// Update one stock dividends
	updateOneStocksDividend := command.UpdateOneStockDividend{}
	bus.Handle(&updateOneStocksDividend, updateOneStockDividendHandler)
	bus.ListenCommand(cbus.AfterSuccess, &updateOneStocksDividend, updateStockDividend)
	bus.ListenCommand(cbus.AfterSuccess, &updateOneStocksDividend, updateStockDividendYield)

	// import transfer
	importTransfer := command.ImportTransfer{}
	bus.Handle(&importTransfer, importTransferHandler)

	// import wallet
	importWallet := command.ImportWallet{}
	bus.Handle(&importWallet, importWalletHandler)

	// import operation
	importOperation := command.ImportOperation{}
	bus.Handle(&importOperation, importOperationHandler)
	bus.ListenCommand(cbus.AfterSuccess, &importOperation, addWalletOperation)
	bus.ListenCommand(cbus.AfterSuccess, &importOperation, updateWalletCapital)

	// List stocks
	listStocks := command.ListStocks{}
	bus.Handle(&listStocks, listStockHandler)

	// Wallet details
	walletDetails := command.WalletDetails{}
	bus.Handle(&walletDetails, walletDetailsHandler)

	// Update wallet stocks price
	updateWalletStocksPrice := command.UpdateWalletStocksPrice{}
	bus.Handle(&updateWalletStocksPrice, updateWalletStocksPriceHandler)
	bus.ListenCommand(cbus.AfterSuccess, &updateWalletStocksPrice, updateStockPrice)
	bus.ListenCommand(cbus.AfterSuccess, &updateWalletStocksPrice, updateStockDividendYield)
	bus.ListenCommand(cbus.AfterSuccess, &updateWalletStocksPrice, updateWalletCapital)
	//bus.ListenCommand(cbus.AfterSuccess, &updateWalletStocksPrice, updateStockPriceVolatility)

	//Reload wallet
	bus.Handle(&command.ReloadWallet{}, reloadWalletHandler)

	// import retention
	importRetention := command.ImportRetention{}
	bus.Handle(&importRetention, importRetentionHandler)

	// add dividend
	addDividend := command.AddDividendOperation{}
	bus.Handle(&addDividend, addOperationHandler)
	bus.ListenCommand(cbus.AfterSuccess, &addDividend, addWalletOperation)
	bus.ListenCommand(cbus.AfterSuccess, &addDividend, updateWalletCapital)
	bus.ListenCommand(cbus.AfterSuccess, &addDividend, registerWalletOperationImport)

	// add buy stock
	addBought := command.AddBuyOperation{}
	bus.Handle(&addBought, addOperationHandler)
	bus.ListenCommand(cbus.AfterSuccess, &addBought, addWalletOperation)
	bus.ListenCommand(cbus.AfterSuccess, &addBought, updateWalletCapital)
	bus.ListenCommand(cbus.AfterSuccess, &addBought, registerWalletOperationImport)

	// add sell stock
	addSold := command.AddSellOperation{}
	bus.Handle(&addSold, addOperationHandler)
	bus.ListenCommand(cbus.AfterSuccess, &addSold, addWalletOperation)
	bus.ListenCommand(cbus.AfterSuccess, &addSold, updateWalletCapital)
	bus.ListenCommand(cbus.AfterSuccess, &addSold, registerWalletOperationImport)

	// add interest
	addInterest := command.AddInterestOperation{}
	bus.Handle(&addInterest, addOperationHandler)
	bus.ListenCommand(cbus.AfterSuccess, &addInterest, addWalletOperation)
	bus.ListenCommand(cbus.AfterSuccess, &addInterest, updateWalletCapital)
	bus.ListenCommand(cbus.AfterSuccess, &addInterest, registerWalletOperationImport)

	// Wallet report
	walletDateDetails := command.WalletDateDetails{}
	bus.Handle(&walletDateDetails, walletDateDetailsHandler)

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
