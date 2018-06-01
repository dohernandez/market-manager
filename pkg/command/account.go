package command

import (
	"context"

	"github.com/urfave/cli"

	exportAccount "github.com/dohernandez/market-manager/pkg/export/account"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// AccountCommand ...
type AccountCommand struct {
	*BaseCommand
	*ExportCommand
}

// NewAccountCommand constructs AccountCommand
func NewAccountCommand(baseCommand *BaseCommand, exportCommand *ExportCommand) *AccountCommand {
	return &AccountCommand{
		BaseCommand:   baseCommand,
		ExportCommand: exportCommand,
	}
}

// ExportWalletItems List in csv format or print into screen the wallet items from a wallet
func (cmd *AccountCommand) ExportWalletItems(cliCtx *cli.Context) error {
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

	ctx = context.WithValue(ctx, "wallet", cliCtx.String("wallet"))
	sorting := cmd.sortingFromCliCtx(cliCtx)

	ex := exportAccount.NewExportWallet(ctx, sorting, c.AccountServiceInstance())
	err = ex.Export()
	if err != nil {
		return err
	}

	return nil
}
