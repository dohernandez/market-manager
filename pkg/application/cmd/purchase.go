package cmd

import (
	"context"

	"github.com/urfave/cli"

	exportPurchase "github.com/dohernandez/market-manager/pkg/application/export/purchase"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
)

type PurchaseCMD struct {
	*BaseCMD
	*BaseExportCMD
}

func NewPurchaseCMD(baseCMD *BaseCMD, baseExportCMD *BaseExportCMD) *PurchaseCMD {
	return &PurchaseCMD{
		BaseCMD:       baseCMD,
		BaseExportCMD: baseExportCMD,
	}
}

// List in csv format the wallet items from a wallet
func (cmd *PurchaseCMD) ExportStocks(cliCtx *cli.Context) error {
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
	ctx = context.WithValue(ctx, "symbol", cliCtx.String("stock"))
	ctx = context.WithValue(ctx, "groupBy", cliCtx.String("group"))
	sorting := cmd.sortingFromCliCtx(cliCtx)

	ex := exportPurchase.NewExportStock(ctx, sorting, c.PurchaseServiceInstance())
	err = ex.Export()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *PurchaseCMD) ExportStocksWithDividend(cliCtx *cli.Context) error {
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
