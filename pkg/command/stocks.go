package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/logger"
)

// StocksCommand ...
type StocksCommand struct {
	*BaseCommand
}

// NewStocksCommand constructs StocksCommand
func NewStocksCommand(baseCommand *BaseCommand) *StocksCommand {
	return &StocksCommand{
		BaseCommand: baseCommand,
	}
}

// Run runs the application stock update
func (cmd *StocksCommand) Run(cliCtx *cli.Context) error {
	return cmd.Price(cliCtx)
}

// Price runs the application stock price update
func (cmd *StocksCommand) Price(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	stockService := c.PurchaseServiceInstance()

	if cliCtx.String("stock") == "" {
		stocks, err := stockService.Stocks()
		if err != nil {
			fmt.Printf("err: %v\n", err)

			return err
		}

		errs := stockService.UpdateLastClosedPriceStocks(stocks)
		if len(errs) > 0 {
			if stocks == nil {
				fmt.Printf("err: %v\n", errs[0])

				return errs[0]
			} else {
				fmt.Printf("some errs happen while updating stocks price: %v\n", errs[0])
			}
		}
	} else {
		stock, err := stockService.FindStockBySymbol(cliCtx.String("stock"))
		if err != nil {
			fmt.Printf("err: %v\n", err)

			return err
		}

		err = stockService.UpdateLastClosedPriceStock(stock)
		if err != nil {
			fmt.Printf("err: %v\n", err)

			return err
		}
	}

	logger.FromContext(ctx).Info("Update finished")

	return nil
}

// Dividend runs the application stock dividend update
func (cmd *StocksCommand) Dividend(cliCtx *cli.Context) error {
	if cliCtx.String("stock") == "" {
		logger.FromContext(context.TODO()).Error("Please specify the stock tricker: market-manager stocks dividend [stock] [file]")

		return nil
	}

	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	stockService := c.PurchaseServiceInstance()

	stk, err := stockService.FindStockBySymbol(cliCtx.String("stock"))
	if err != nil {
		fmt.Printf("err: %v\n", err)

		return err
	}
	fmt.Printf("%+v\n", stk)

	return nil
}
