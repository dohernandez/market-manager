package command

import (
	"context"

	"github.com/urfave/cli"

	"errors"

	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/jmoiron/sqlx"
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

// Run runs the application import data
func (cmd *ImportCommand) Run(cliCtx *cli.Context) error {
	if cliCtx.String("type") == "" {
		logger.FromContext(context.TODO()).Error("Please specify the import type: market-manager [type] [file]")

		return nil
	}

	if cliCtx.String("file") == "" {
		logger.FromContext(context.TODO()).Error("Please specify the import file: market-manager [type] [file]")

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

	i, err := cmd.getImport(cliCtx, ctx, db, cliCtx.String("type"), cliCtx.String("file"))
	if err != nil {
		logger.FromContext(context.TODO()).WithError(err).Error("Failed importing")

		return nil
	}

	err = i.Import()
	if err != nil {
		logger.FromContext(context.TODO()).WithError(err).Error("Failed importing")

		return nil
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}

func (cmd *ImportCommand) getImport(cliCtx *cli.Context, ctx context.Context, db *sqlx.DB, t, file string) (_import.Import, error) {
	c := cmd.Container(db)
	r := _import.NewCsvReader(file)

	switch t {
	case "stock":
		return _import.NewImportStock(ctx, r, c.MarketFinderInstance(), c.ExchangeFinderInstance(), c.StockServiceInstance()), nil
	case "dividend":
		if cliCtx.String("stock") != "" {
			ctx = context.WithValue(ctx, "stock", cliCtx.String("stock"))
		}

		if cliCtx.String("status") != "" {
			ctx = context.WithValue(ctx, "status", cliCtx.String("status"))
		}

		return _import.NewImportStockDividend(ctx, r, c.StockServiceInstance()), nil
	}

	return nil, errors.New("type not supported")
}
