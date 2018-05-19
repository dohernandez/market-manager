package command

import (
	"context"

	"github.com/urfave/cli"

	"fmt"

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

// Run runs the application import data
func (cmd *StocksCommand) Run(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	stockService := c.StockServiceInstance()

	stocks, errs := stockService.UpdateLastClosedPriceToAllStocks()

	if len(errs) > 0 {
		if stocks == nil {
			fmt.Printf("err: %v\n", errs[0])

			return errs[0]
		} else {
			fmt.Printf("some errs happen while updating stocks price: %v\n", errs[0])
		}
	}

	for _, stock := range stocks {
		fmt.Printf("%+v\n", stock)
	}

	return nil
}
