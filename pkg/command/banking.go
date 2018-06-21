package command

import (
	"context"
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/container"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/import/banking"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// BankingCommand ...
type BankingCommand struct {
	*BaseCommand
	*BaseImportCommand
}

// NewBankingCommand constructs BankingCommand
func NewBankingCommand(baseCommand *BaseCommand, baseImportCommand *BaseImportCommand) *BankingCommand {
	return &BankingCommand{
		BaseCommand:       baseCommand,
		BaseImportCommand: baseImportCommand,
	}
}

func (cmd *BankingCommand) ImportTransfer(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	tis, err := cmd.getTransferImport(cliCtx, cmd.config.Import.TransfersPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	err = cmd.runImport(ctx, c, "transfers", tis, func(ctx context.Context, c *container.Container, ri resourceImport) error {
		r := _import.NewCsvReader(ri.filePath)
		i := import_banking.NewImportTransfer(ctx, r, c.BankingServiceInstance())

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
func (cmd *BankingCommand) getTransferImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
	var ris []resourceImport

	if cliCtx.String("file") != "" {
		ris = append(ris, resourceImport{
			filePath:     cliCtx.String("file"),
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
