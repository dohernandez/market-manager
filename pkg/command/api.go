package command

import (
	"context"
	"net/http"

	"fmt"

	"github.com/dohernandez/market-manager/pkg/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/urfave/cli"
)

// ApiCommand ...
type ApiCommand struct {
	*BaseCommand
}

// NewApiCommand constructs ApiCommand
func NewApiCommand(baseCommand *BaseCommand) *ApiCommand {
	return &ApiCommand{
		BaseCommand: baseCommand,
	}
}

// Run runs the application import data
func (cmd *ApiCommand) Run(cliCtx *cli.Context) error {
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
	stks, err := stockService.Stocks()
	if err != nil {
		fmt.Printf("%+v\n", err)

		return err
	}

	for _, stk := range stks {
		client := iex.NewClient(&http.Client{})

		q, err := client.Quote.Get(stk.Symbol)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}

		fmt.Printf("%s %+v\n", stk.Symbol, q)
	}

	return nil
}
