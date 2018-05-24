package command

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"path/filepath"

	"os"

	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/import/account"
	"github.com/dohernandez/market-manager/pkg/import/banking"
	"github.com/dohernandez/market-manager/pkg/import/purchase"
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

	if cliCtx.String("stock") == "" {
		logger.FromContext(ctx).Fatal("Please specify the stock: market-manager stocks import dividend [stock]")
	}

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
		file := fmt.Sprintf("%s/%s_%s.csv", cmd.config.Import.DividendsPath, strings.ToLower(cliCtx.String("stock")), st)
		r := _import.NewCsvReader(file)
		i := import_purchase.NewImportStockDividend(ctx, r, c.PurchaseServiceInstance())

		err = i.Import()
		if err != nil {
			logger.FromContext(ctx).WithError(err).Error("Failed importing")

			return err
		}
	}

	logger.FromContext(ctx).Info("Import finished")

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

	type walletImport struct {
		file   string
		wallet string
	}
	var wis []walletImport

	if cliCtx.String("file") == "" && cliCtx.String("wallet") != "" {
		wallet := cliCtx.String("wallet")
		file := fmt.Sprintf("%s/%s.csv", cmd.config.Import.WalletSPath, wallet)

		wis = append(wis, walletImport{
			file:   file,
			wallet: wallet,
		})
	} else if cliCtx.String("wallet") == "" && cliCtx.String("file") != "" {
		file := cliCtx.String("file")
		wallet := cmd.getWalletFromFile(file)

		wis = append(wis, walletImport{
			file:   file,
			wallet: wallet,
		})
	} else {
		err := filepath.Walk(cmd.config.Import.WalletSPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".csv" {
				file := path
				wallet := cmd.getWalletFromFile(file)
				wis = append(wis, walletImport{
					file:   file,
					wallet: wallet,
				})
			}

			return nil
		})
		if err != nil {
			panic(err)
		}
	}

	for _, wi := range wis {
		ctx = context.WithValue(ctx, "name", wi.wallet)

		r := _import.NewCsvReader(wi.file)
		i := import_account.NewImportWallet(ctx, r, c.AccountServiceInstance(), c.BankingServiceInstance())

		err = i.Import()
		if err != nil {
			logger.FromContext(ctx).WithError(err).Error("Failed importing")

			return err
		}
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}

func (cmd *ImportCommand) getWalletFromFile(file string) string {
	var dir = filepath.Dir(file)
	var ext = filepath.Ext(file)
	return file[len(dir)+1 : len(file)-len(ext)]
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

	file := cliCtx.String("file")
	if cliCtx.String("file") == "" {
		file = fmt.Sprintf("%s/account.csv", cmd.config.Import.StocksPath)
	}

	r := _import.NewCsvReader(file)
	i := import_account.NewImportAccount(ctx, r, c.PurchaseServiceInstance(), c.AccountServiceInstance())

	err = i.Import()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Error("Failed importing")

		return err
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
		logger.FromContext(ctx).WithError(err).Error("Failed importing")

		return err
	}

	logger.FromContext(ctx).Info("Import finished")

	return nil
}
