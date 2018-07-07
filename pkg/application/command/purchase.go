package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	app "github.com/dohernandez/market-manager/pkg/application"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/import/purchase"
	"github.com/dohernandez/market-manager/pkg/logger"
)

type PurchaseCommand struct {
	*BaseCommand
	*BaseImportCommand
	*BaseExportCommand
}

func NewPurchaseCommand(baseCommand *BaseCommand, baseImportCommand *BaseImportCommand, baseExportCommand *BaseExportCommand) *PurchaseCommand {
	return &PurchaseCommand{
		BaseCommand:       baseCommand,
		BaseImportCommand: baseImportCommand,
		BaseExportCommand: baseExportCommand,
	}
}

func (cmd *PurchaseCommand) ImportStock(cliCtx *cli.Context) error {
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

	err = cmd.runImport(ctx, c, "stocks", sis, func(ctx context.Context, c *app.Container, ri resourceImport) error {
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

func (cmd *PurchaseCommand) getStockImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
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
