package command

import (
	"context"

	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/export"
	exportAccount "github.com/dohernandez/market-manager/pkg/export/account"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// AccountCommand ...
type AccountCommand struct {
	*BaseCommand
}

// NewAccountCommand constructs AccountCommand
func NewAccountCommand(baseCommand *BaseCommand) *AccountCommand {
	return &AccountCommand{
		BaseCommand: baseCommand,
	}
}

// List in csv format the wallet items from a wallet
func (cmd *AccountCommand) WalletItems(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).WithError(err).Fatal("Missing wallet name")
	}

	sortBy := exportAccount.Stock
	orderBy := export.Descending

	if cliCtx.String("sort") != "" {
		sortBy = export.SortBy(cliCtx.String("sort"))
	}
	if cliCtx.String("order") != "" {
		orderBy = export.OrderBy(cliCtx.String("order"))
	}

	ctx = context.WithValue(ctx, "wallet", cliCtx.String("wallet"))
	sorting := export.Sorting{
		By:    sortBy,
		Order: orderBy,
	}

	ex := exportAccount.NewExportWallet(ctx, sorting, c.AccountServiceInstance())
	err = ex.Export()
	if err != nil {
		return err
	}

	return nil
}
