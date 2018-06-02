package command

import (
	"context"

	"github.com/urfave/cli"

	"fmt"
	"os"
	"path/filepath"

	"github.com/dohernandez/market-manager/pkg/container"
	exportAccount "github.com/dohernandez/market-manager/pkg/export/account"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/import/account"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// AccountCommand ...
type AccountCommand struct {
	*BaseCommand
	*ImportCommand
	*ExportCommand
}

// NewAccountCommand constructs AccountCommand
func NewAccountCommand(baseCommand *BaseCommand, importCommand *ImportCommand, exportCommand *ExportCommand) *AccountCommand {
	return &AccountCommand{
		BaseCommand:   baseCommand,
		ImportCommand: importCommand,
		ExportCommand: exportCommand,
	}
}

// ImportWallet
func (cmd *AccountCommand) ImportWallet(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	wis, err := cmd.getWalletImport(cliCtx, cmd.config.Import.WalletsPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	cmd.runImport(ctx, c, "wallets", wis, func(ctx context.Context, c *container.Container, ri resourceImport) error {
		ctx = context.WithValue(ctx, "wallet", ri.resourceName)

		r := _import.NewCsvReader(ri.filePath)
		i := import_account.NewImportWallet(ctx, r, c.AccountServiceInstance(), c.BankingServiceInstance())

		err = i.Import()
		if err != nil {
			logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
		}

		return nil
	})

	logger.FromContext(ctx).Info("Import finished")

	return nil
}

func (cmd *AccountCommand) getWalletImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
	var wis []resourceImport

	if cliCtx.String("file") == "" && cliCtx.String("wallet") != "" {
		walletName := cliCtx.String("wallet")
		filePath := fmt.Sprintf("%s/%s.csv", importPath, walletName)

		wis = append(wis, resourceImport{
			filePath:     filePath,
			resourceName: walletName,
		})
	} else if cliCtx.String("wallet") == "" && cliCtx.String("file") != "" {
		filePath := cliCtx.String("file")
		walletName := cmd.geResourceNameFromFilePath(filePath)

		wis = append(wis, resourceImport{
			filePath:     filePath,
			resourceName: walletName,
		})
	} else {
		err := filepath.Walk(importPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".csv" {
				filePath := path
				walletName := cmd.geResourceNameFromFilePath(filePath)
				wis = append(wis, resourceImport{
					filePath:     filePath,
					resourceName: walletName,
				})
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return wis, nil
}

func (cmd *AccountCommand) ImportOperation(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	ois, err := cmd.getWalletImport(cliCtx, cmd.config.Import.AccountsPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	for _, oi := range ois {
		ctx = context.WithValue(ctx, "wallet", oi.resourceName)

		r := _import.NewCsvReader(oi.filePath)
		i := import_account.NewImportAccount(ctx, r, c.PurchaseServiceInstance(), c.AccountServiceInstance())

		err = i.Import()
		if err != nil {
			logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
		}
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
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
