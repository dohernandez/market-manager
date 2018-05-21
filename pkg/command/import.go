package command

import (
	"context"

	"github.com/urfave/cli"

	"errors"

	"fmt"

	"strings"

	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// ImportCommand ...
type ImportCommand struct {
	*BaseCommand
}

// NewImportCommand constructs ImportCommand
func NewImportCommand(baseCommand *BaseCommand) *ImportCommand {
	return &ImportCommand{
		BaseCommand: baseCommand,
	}
}

// RunImportQuote runs the application import data
func (cmd *ImportCommand) RunImportQuote(cliCtx *cli.Context) error {
	if cliCtx.String("file") == "" {
		logger.FromContext(context.TODO()).Fatal("Please specify the import file: market-manager [type] [file]")
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

	file := fmt.Sprintf("%s/stocks.csv", cmd.config.QUOTE.StocksPath)
	r := _import.NewCsvReader(file)
	i := _import.NewImportStock(ctx, r, c.MarketFinderInstance(), c.ExchangeFinderInstance(), c.StockServiceInstance())

	err = i.Import()
	if err != nil {
		logger.FromContext(context.TODO()).WithError(err).Error("Failed importing")

		return err
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}

// RunImportDividend runs the application import data
func (cmd *ImportCommand) RunImportDividend(cliCtx *cli.Context) error {
	if cliCtx.String("stock") == "" {
		logger.FromContext(context.TODO()).Fatal("Please specify the stock: market-manager stocks import dividend [stock]")
	}

	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	ctx = context.WithValue(ctx, "stock", cliCtx.String("stock"))

	var status []string
	if cliCtx.String("status") != "" {
		if cliCtx.String("status") != "payed" || cliCtx.String("status") != "projected" {
			return errors.New("invalid status value. Status values allow [projected]")
		}

		status = append(status, cliCtx.String("status"))
	} else {
		status = append(status, "payed")
		status = append(status, "projected")
	}

	c := cmd.Container(db)

	for _, st := range status {
		ctx = context.WithValue(ctx, "status", st)
		file := fmt.Sprintf("%s/%s_%s.csv", cmd.config.QUOTE.DividendsPath, strings.ToLower(cliCtx.String("stock")), st)
		r := _import.NewCsvReader(file)
		i := _import.NewImportStockDividend(ctx, r, c.StockServiceInstance())

		err = i.Import()
		if err != nil {
			logger.FromContext(context.TODO()).WithError(err).Error("Failed importing")

			return err
		}
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}
