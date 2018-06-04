package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	"os"
	"path/filepath"

	"strings"

	"github.com/dohernandez/market-manager/pkg/container"
	exportPurchase "github.com/dohernandez/market-manager/pkg/export/purchase"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/import/purchase"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// StocksCommand ...
type StocksCommand struct {
	*BaseCommand
	*ImportCommand
	*ExportCommand
}

// NewStocksCommand constructs StocksCommand
func NewStocksCommand(baseCommand *BaseCommand, importCommand *ImportCommand, exportCommand *ExportCommand) *StocksCommand {
	return &StocksCommand{
		BaseCommand:   baseCommand,
		ImportCommand: importCommand,
		ExportCommand: exportCommand,
	}
}

// Quote runs the application import data
func (cmd *StocksCommand) ImportStock(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	sis, err := cmd.getStockImport(cliCtx, cmd.config.Import.StocksPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	err = cmd.runImport(ctx, c, "stocks", sis, func(ctx context.Context, c *container.Container, ri resourceImport) error {
		ctx = context.WithValue(ctx, "stock", ri.resourceName)

		r := _import.NewCsvReader(ri.filePath)
		i := import_purchase.NewImportStock(ctx, r, c.PurchaseServiceInstance())

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

func (cmd *StocksCommand) getStockImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
	var ris []resourceImport

	if cliCtx.String("file") != "" {
		filePath := fmt.Sprintf("%s/%s.csv", importPath, cliCtx.String("file"))

		ris = append(ris, resourceImport{
			filePath:     filePath,
			resourceName: "",
		})
	} else {
		err := filepath.Walk(importPath, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".csv" {
				filePath := path
				ris = append(ris, resourceImport{
					filePath:     filePath,
					resourceName: "",
				})
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return ris, nil
}

// Price runs the application stock price update
func (cmd *StocksCommand) Price(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	stockService := c.PurchaseServiceInstance()

	if cliCtx.String("stock") == "" {
		stocks, err := stockService.Stocks()
		if err != nil {
			logger.FromContext(ctx).Debugf("err: %v\n", err)

			return err
		}

		errs := stockService.UpdateLastClosedPriceStocks(stocks)
		if len(errs) > 0 {
			if stocks == nil {
				for _, err := range errs {
					logger.FromContext(ctx).Debugf("err: %v\n", err)
				}

				return errs[0]
			} else {
				logger.FromContext(ctx).Debug("some errs happen while updating stocks price:")
				for _, err := range errs {
					logger.FromContext(ctx).Debugf("err: %v\n", err)
				}
			}
		}
	} else {
		stock, err := stockService.FindStockBySymbol(cliCtx.String("stock"))
		if err != nil {
			logger.FromContext(ctx).Debugf("err: %v\n", err)

			return err
		}

		err = stockService.UpdateLastClosedPriceStock(stock)
		if err != nil {
			logger.FromContext(ctx).Debug("some errs happen while updating stocks price:")
			logger.FromContext(ctx).Debugf("err: %v\n", err)

			return err
		}
	}

	logger.FromContext(ctx).Info("Update finished")

	return nil
}

// ImportDividend runs the application import data
func (cmd *StocksCommand) ImportDividend(cliCtx *cli.Context) error {
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

	err = cmd.runImport(ctx, c, "dividends", sdis, func(ctx context.Context, c *container.Container, ri resourceImport) error {
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

func (cmd *StocksCommand) getStockDividendImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
	var ris []resourceImport

	stockName := cliCtx.String("stock")
	if cliCtx.String("file") != "" && cliCtx.String("stock") != "" {
		filePath := fmt.Sprintf("%s/%s.csv", importPath, cliCtx.String("file"))

		ris = append(ris, resourceImport{
			filePath:     filePath,
			resourceName: cmd.sanitizeStockName(stockName),
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
						resourceName: cmd.sanitizeStockName(stockNameFromFile),
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

func (cmd *StocksCommand) sanitizeStockName(stockName string) string {
	return strings.Replace(stockName, ":", ".", 1)
}

// Dividend runs the application stock dividend update
func (cmd *StocksCommand) Dividend(cliCtx *cli.Context) error {
	if cliCtx.String("stock") == "" {
		logger.FromContext(context.TODO()).Error("Please specify the stock tricker: market-manager stocks dividend [stock] [file]")

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

	c := cmd.Container(db)

	stockService := c.PurchaseServiceInstance()

	stk, err := stockService.FindStockBySymbol(cliCtx.String("stock"))
	if err != nil {
		fmt.Printf("err: %v\n", err)

		return err
	}
	fmt.Printf("%+v\n", stk)

	return nil
}

// List in csv format the wallet items from a wallet
func (cmd *StocksCommand) ExportStocks(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	ctx = context.WithValue(ctx, "exchange", cliCtx.String("exchange"))
	sorting := cmd.sortingFromCliCtx(cliCtx)

	ex := exportPurchase.NewExportStock(ctx, sorting, c.PurchaseServiceInstance())
	err = ex.Export()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *StocksCommand) ExportStocksWithDividend(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	ctx = context.WithValue(ctx, "year", cliCtx.String("year"))
	ctx = context.WithValue(ctx, "month", cliCtx.String("month"))
	sorting := cmd.sortingFromCliCtx(cliCtx)

	ex := exportPurchase.NewExportStockWithDividends(ctx, sorting, c.PurchaseServiceInstance())
	err = ex.Export()
	if err != nil {
		return err
	}

	return nil
}
