package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"path/filepath"

	"os"

	"regexp"

	"path"

	"github.com/dohernandez/market-manager/pkg/container"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/import/account"
	"github.com/dohernandez/market-manager/pkg/import/banking"
	"github.com/dohernandez/market-manager/pkg/import/purchase"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	// ImportCommand ...
	ImportCommand struct {
		*BaseCommand
	}

	resourceImport struct {
		filePath     string
		resourceName string
	}
)

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
		file = fmt.Sprintf("%s/stocks.csv", cmd.config.Import.StocksPath)
	}

	r := _import.NewCsvReader(file)
	i := import_purchase.NewImportStock(ctx, r, c.PurchaseServiceInstance())

	err = i.Import()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Error("Failed importing")

		return err
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}

// Dividend runs the application import data
func (cmd *ImportCommand) Dividend(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	sdis, err := cmd.getStockDividendImport(cliCtx, cmd.config.Import.DividendsPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	err = runImport(ctx, c, "dividends", sdis, func(ctx context.Context, c *container.Container, ri resourceImport) error {
		ctx = context.WithValue(ctx, "stock", ri.resourceName)

		r := _import.NewCsvReader(ri.filePath)
		i := import_purchase.NewImportStockDividend(ctx, r, c.PurchaseServiceInstance())

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

func (cmd *ImportCommand) getStockDividendImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
	var ris []resourceImport

	stockName := cliCtx.String("stock")
	if cliCtx.String("file") != "" && cliCtx.String("stock") != "" {
		filePath := fmt.Sprintf("%s/%s.csv", importPath, cliCtx.String("file"))

		ris = append(ris, resourceImport{
			filePath:     filePath,
			resourceName: stockName,
		})
	} else {
		err := filepath.Walk(importPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".csv" {
				filePath := path
				stockNameFromFile := cmd.geResourceNameFromFilePath(filePath)
				if len(stockName) == 0 || stockName == stockNameFromFile {
					ris = append(ris, resourceImport{
						filePath:     filePath,
						resourceName: stockNameFromFile,
					})
				}
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return ris, nil
}

func runImport(
	ctx context.Context,
	c *container.Container,
	resourceType string,
	ris []resourceImport,
	fn func(ctx context.Context, c *container.Container, ri resourceImport) error,
) error {
	is := c.ImportStorageInstance()
	irs, err := is.FindAllByResource(resourceType)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		irs = []_import.Resource{}
	}

	for _, ri := range ris {
		fileName := path.Base(ri.filePath)

		var found bool
		for _, ir := range irs {
			if ir.FileName == fileName {
				found = true

				break
			}
		}

		if !found {
			if err := fn(ctx, c, ri); err != nil {
				return err
			}

			ir := _import.NewResource(resourceType, fileName)
			err := is.Persist(ir)
			if err != nil {
				return err
			}

			logger.FromContext(ctx).Infof("Imported file %s", fileName)
		}
	}

	return nil
}

func (cmd *ImportCommand) Wallet(cliCtx *cli.Context) error {
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

	runImport(ctx, c, "wallets", wis, func(ctx context.Context, c *container.Container, ri resourceImport) error {
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

func (cmd *ImportCommand) getWalletImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
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

func (cmd *ImportCommand) geResourceNameFromFilePath(file string) string {
	var dir = filepath.Dir(file)
	var ext = filepath.Ext(file)

	name := file[len(dir)+1 : len(file)-len(ext)]

	reg := regexp.MustCompile(`(^[0-9]+_)+(.*)`)
	res := reg.ReplaceAllString(name, "${2}")

	return res
}

func (cmd *ImportCommand) Operation(cliCtx *cli.Context) error {
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

func (cmd *ImportCommand) Transfer(cliCtx *cli.Context) error {
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
		file = fmt.Sprintf("%s/transfers.csv", cmd.config.Import.TransfersPath)
	}

	r := _import.NewCsvReader(file)
	i := import_banking.NewImportTransfer(ctx, r, c.BankingServiceInstance())

	err = i.Import()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}
