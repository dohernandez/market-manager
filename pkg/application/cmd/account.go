package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/application"
	exportAccount "github.com/dohernandez/market-manager/pkg/application/export/account"
	"github.com/dohernandez/market-manager/pkg/application/import"
	"github.com/dohernandez/market-manager/pkg/application/import/account"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// AccountCMD ...
type AccountCMD struct {
	*BaseCMD
	*BaseImportCMD
	*BaseExportCMD
}

// NewAccountCMD constructs AccountCMD
func NewAccountCMD(baseCMD *BaseCMD, baseImportCMD *BaseImportCMD, baseExportCMD *BaseExportCMD) *AccountCMD {
	return &AccountCMD{
		BaseCMD:       baseCMD,
		BaseImportCMD: baseImportCMD,
		BaseExportCMD: baseExportCMD,
	}
}

// ImportWallet
func (cmd *AccountCMD) ImportWallet(cliCtx *cli.Context) error {
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

	cmd.runImport(ctx, c, "wallets", wis, func(ctx context.Context, c *app.Container, ri resourceImport) error {
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

func (cmd *AccountCMD) getWalletImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
	var wis []resourceImport

	if cliCtx.String("file") == "" && cliCtx.String("wallet") != "" {
		walletName := cliCtx.String("wallet")

		err := filepath.Walk(importPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".csv" {
				filePath := path
				wName := cmd.geResourceNameFromFilePath(filePath)

				if wName == walletName {
					wis = append(wis, resourceImport{
						filePath:     filePath,
						resourceName: wName,
					})
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

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

func (cmd *AccountCMD) ImportOperation(cliCtx *cli.Context) error {
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

	err = cmd.runImport(ctx, c, "accounts", ois, func(ctx context.Context, c *app.Container, ri resourceImport) error {
		ctx = context.WithValue(ctx, "wallet", ri.resourceName)

		r := _import.NewCsvReader(ri.filePath)
		i := import_account.NewImportAccount(ctx, r, c.PurchaseServiceInstance(), c.AccountServiceInstance())

		err = i.Import()
		if err != nil {
			logger.FromContext(ctx).WithError(err).Fatal("Failed importing %s", ri.filePath)
		}

		return nil
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
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
