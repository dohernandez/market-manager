package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gogolfing/cbus"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // enable postgres driver
	"github.com/sirupsen/logrus"

	"github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/handler"
	"github.com/dohernandez/market-manager/pkg/application/listener"
	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/application/storage"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
)

func main() {
	// create context
	ctxt, cancel := context.WithCancel(context.Background())

	// Init logger
	log := logrus.StandardLogger()
	level, err := logrus.ParseLevel("debug")
	if err != nil {
		log.WithError(err).Fatal("Parsing LogLevel fails.")
	}
	log.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{FieldMap: logrus.FieldMap{
		logrus.FieldKeyTime: "@timestamp",
		logrus.FieldKeyMsg:  "message",
	}})

	logger.SetLogger(log)

	// Init postgres connection
	db, err := sqlx.Connect("postgres", "postgres://mms:mms4you@127.0.0.1:4532/market-manager-service?sslmode=disable&client_encoding=UTF8")
	if err != nil {
		log.WithError(err).Fatal(err, "Connecting to postgres fails")
	}

	// Init chromedp
	cdp, err := chromedp.New(ctxt)
	if err != nil {
		log.WithError(err).Fatal("Init chromedp fails.")
	}
	defer func() {
		shutdown(ctxt, cdp, cancel)
	}()

	handleGracefulShutdown(ctxt, cdp, cancel)

	// FINDER
	stockFinder := storage.NewStockFinder(db)
	walletFinder := storage.NewWalletFinder(db)
	stockDividendFinder := storage.NewStockDividendFinder(db)

	// PERSISTER
	stockDividendPersister := storage.NewStockDividendPersister(db)
	stockPersister := storage.NewStockPersister(db)

	// SCRAPER
	marketChameleonWWWUrlBuilder := service.NewStockScrapeMarketChameleonWWWUrlBuilder(ctxt, "https://marketchameleon.com/Overview")
	marketChameleonWWWHtmlParser := service.NewStockDividendMarketChameleonChromedpParser(ctxt, cdp)

	// SERVICE
	stockDividendMarketChameleonService := service.NewStockDividendMarketChameleon(ctxt, marketChameleonWWWUrlBuilder, marketChameleonWWWHtmlParser)

	// HANDLER
	updateWalletStocksDividendHandler := handler.NewUpdateWalletStocksDividend(walletFinder, stockFinder)

	// LISTENER
	updateStockDividend := listener.NewUpdateStockDividend(stockDividendPersister, stockDividendMarketChameleonService)
	updateStockDividend.WithConcurrency(1)
	updateStockDividend.WithSleep(5)
	updateStockDividendYield := listener.NewUpdateStockDividendYield(stockDividendFinder, stockPersister)
	// COMMAND BUS
	bus := cbus.Bus{}

	// Update wallet stock dividends
	updateWalletStocksDividend := command.UpdateWalletStocksDividend{}
	bus.Handle(&updateWalletStocksDividend, updateWalletStocksDividendHandler)
	bus.ListenCommand(cbus.AfterSuccess, &updateWalletStocksDividend, updateStockDividend)
	bus.ListenCommand(cbus.AfterSuccess, &updateWalletStocksDividend, updateStockDividendYield)

	walletName := "degiro"
	_, err = bus.ExecuteContext(ctxt, &command.UpdateWalletStocksDividend{Wallet: walletName})
	if err != nil {
		logger.FromContext(ctxt).WithError(err).Error("fail")
	}

	logger.FromContext(ctxt).Info("finish successful")
}

func handleGracefulShutdown(ctxt context.Context, c *chromedp.CDP, cancel func()) {
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGTERM, os.Interrupt)

	go func() {
		<-exit

		gracefulDelay := 20 * time.Second
		// Killing with no mercy after a graceful delay
		go func() {
			time.Sleep(gracefulDelay)

			log.Fatalf("Failed to gracefully shutdown server in %s, exiting.", gracefulDelay)

			os.Exit(0)
		}()

		log.Println("Graceful shutdown")
		shutdown(ctxt, c, cancel)
	}()
}

func shutdown(ctxt context.Context, c *chromedp.CDP, cancel func()) {
	log.Println("Shutting down ...")

	// shutdown chrome
	err := c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	if cancel != nil {
		cancel()
	}
}
