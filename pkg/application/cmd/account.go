package cmd

import (
	"context"

	"github.com/urfave/cli"

	exportAccount "github.com/dohernandez/market-manager/pkg/application/export/account"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
)

// AccountCMD ...
type AccountCMD struct {
	*BaseCMD
	*BaseExportCMD
}

// NewAccountCMD constructs AccountCMD
func NewAccountCMD(baseCMD *BaseCMD, baseExportCMD *BaseExportCMD) *AccountCMD {
	return &AccountCMD{
		BaseCMD:       baseCMD,
		BaseExportCMD: baseExportCMD,
	}
}

// ExportWalletItems List in csv format or print into screen the wallet items from a wallet
func (cmd *AccountCMD) ExportWalletItems(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	ctx = context.WithValue(ctx, "wallet", cliCtx.String("wallet"))
	ctx = context.WithValue(ctx, "stock", cliCtx.String("stock"))
	ctx = context.WithValue(ctx, "sells", cliCtx.String("sells"))
	ctx = context.WithValue(ctx, "buys", cliCtx.String("buys"))
	sorting := cmd.sortingFromCliCtx(cliCtx)

	ex := exportAccount.NewExportWallet(ctx, sorting, cmd.config, c.AccountServiceInstance(), c.CurrencyConverterClientInstance())
	return cmd.runExport(ex)
}
