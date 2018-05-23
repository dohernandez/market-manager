package command

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/import/account"
	"github.com/dohernandez/market-manager/pkg/import/market"
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

// Quote runs the application import data
func (cmd *ImportCommand) Quote(cliCtx *cli.Context) error {

	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	file := cliCtx.String("file")
	if cliCtx.String("file") == "" {
		file = fmt.Sprintf("%s/stocks.csv", cmd.config.QUOTE.StocksPath)
	}

	r := _import.NewCsvReader(file)
	i := import_market.NewImportStock(ctx, r, c.PurchaseServiceInstance())

	err = i.Import()
	if err != nil {
		logger.FromContext(context.TODO()).WithError(err).Error("Failed importing")

		return err
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}

// Dividend runs the application import data
func (cmd *ImportCommand) Dividend(cliCtx *cli.Context) error {
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
		i := import_market.NewImportStockDividend(ctx, r, c.PurchaseServiceInstance())

		err = i.Import()
		if err != nil {
			logger.FromContext(context.TODO()).WithError(err).Error("Failed importing")

			return err
		}
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}

func (cmd *ImportCommand) Account(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	file := cliCtx.String("file")
	if cliCtx.String("file") == "" {
		file = fmt.Sprintf("%s/account.csv", cmd.config.QUOTE.StocksPath)
	}

	r := _import.NewCsvReader(file)
	i := import_account.NewImportAccount(ctx, r, c.PurchaseServiceInstance(), c.AccountServiceInstance())

	err = i.Import()
	if err != nil {
		logger.FromContext(context.TODO()).WithError(err).Error("Failed importing")

		return err
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}
