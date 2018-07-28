package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/gogolfing/cbus"
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/render"
	"github.com/dohernandez/market-manager/pkg/application/storage"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	resourceImport struct {
		filePath     string
		resourceName string
	}

	CLI struct {
		*Base
		resourceStorage util.ResourceStorage
	}
)

func NewCLI(base *Base) *CLI {
	return &CLI{
		Base:            base,
		resourceStorage: storage.NewUtilImportStorage(base.DB),
	}
}

func (cmd *CLI) UpdatePrice(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	if cliCtx.String("stock") == "" {
		_, err := bus.ExecuteContext(ctx, &command.UpdateAllStocksPrice{})
		if err != nil {
			return err
		}
	} else {
		_, err := bus.ExecuteContext(ctx, &command.UpdateOneStockPrice{Symbol: cliCtx.String("stock")})
		if err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Update finished")

	return nil
}

func (cmd *CLI) ImportStock(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	ris, err := cmd.getStockImport(cliCtx, cmd.config.Import.StocksPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	err = cmd.runImport(ctx, bus, "stocks", ris, func(ctx context.Context, bus *cbus.Bus, ri resourceImport) error {
		_, err := bus.ExecuteContext(ctx, &command.ImportStock{FilePath: ri.filePath})
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

func (cmd *CLI) runImport(
	ctx context.Context,
	bus *cbus.Bus,
	resourceType string,
	ris []resourceImport,
	fn func(ctx context.Context, bus *cbus.Bus, ri resourceImport) error,
) error {
	irs, err := cmd.resourceStorage.FindAllByResource(resourceType)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		irs = []util.Resource{}
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
			logger.FromContext(ctx).Infof("Importing file %s", fileName)

			if err := fn(ctx, bus, ri); err != nil {
				return err
			}

			ir := util.NewResource(resourceType, fileName)
			err := cmd.resourceStorage.Persist(ir)
			if err != nil {
				return err
			}

			logger.FromContext(ctx).Infof("Imported file %s", fileName)
		}
	}

	return nil
}

func (cmd *CLI) geResourceNameFromFilePath(file string) string {
	var dir = filepath.Dir(file)
	var ext = filepath.Ext(file)

	name := file[len(dir)+1 : len(file)-len(ext)]

	reg := regexp.MustCompile(`(^[0-9]+_)+(.*)`)
	res := reg.ReplaceAllString(name, "${2}")

	return res
}

func (cmd *CLI) getStockImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
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

func (cmd *CLI) UpdateDividend(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	if cliCtx.String("stock") == "" {
		_, err := bus.ExecuteContext(ctx, &command.UpdateAllStockDividend{})
		if err != nil {
			return err
		}
	} else {
		_, err := bus.ExecuteContext(ctx, &command.UpdateOneStockDividend{
			Symbol: cliCtx.String("stock"),
		})
		if err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Update finished")

	return nil
}

func (cmd *CLI) ImportTransfer(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	ris, err := cmd.getTransferImport(cliCtx, cmd.config.Import.TransfersPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	err = cmd.runImport(ctx, bus, "transfers", ris, func(ctx context.Context, bus *cbus.Bus, ri resourceImport) error {
		_, err := bus.ExecuteContext(ctx, &command.ImportTransfer{FilePath: ri.filePath})
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

func (cmd *CLI) getTransferImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
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

// ImportWallet
func (cmd *CLI) ImportWallet(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	wis, err := cmd.getWalletImport(cliCtx, cmd.config.Import.WalletsPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	err = cmd.runImport(ctx, bus, "wallets", wis, func(ctx context.Context, bus *cbus.Bus, ri resourceImport) error {
		_, err := bus.ExecuteContext(ctx, &command.ImportWallet{FilePath: ri.filePath, Name: ri.resourceName})
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

func (cmd *CLI) getWalletImport(cliCtx *cli.Context, importPath string) ([]resourceImport, error) {
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

func (cmd *CLI) ImportOperation(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	ois, err := cmd.getWalletImport(cliCtx, cmd.config.Import.AccountsPath)
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed importing")
	}

	err = cmd.runImport(ctx, bus, "accounts", ois, func(ctx context.Context, bus *cbus.Bus, ri resourceImport) error {
		_, err := bus.ExecuteContext(ctx, &command.ImportOperation{FilePath: ri.filePath, Wallet: ri.resourceName})
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

// List in csv format the wallet items from a wallet
func (cmd *CLI) ExportStocks(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	stks, err := bus.ExecuteContext(ctx, &command.ListStocks{
		Exchange: cliCtx.String("exchange"),
	})
	if err != nil {
		return err
	}

	sls := render.NewScreenListStocks(2)
	sls.Render(&render.OutputScreenListStocks{
		Stocks:  stks.([]*render.StockOutput),
		GroupBy: util.GroupBy(cliCtx.String("group")),
		Sorting: cmd.sortingFromCliCtx(cliCtx),
	})

	return nil
}

func (cmd *CLI) sortingFromCliCtx(cliCtx *cli.Context) util.Sorting {
	sortBy := util.SortByStock
	orderBy := util.OrderDescending

	if cliCtx.String("sort") != "" {
		sortBy = util.SortBy(cliCtx.String("sort"))
	}
	if cliCtx.String("order") != "" {
		orderBy = util.OrderBy(cliCtx.String("order"))
	}

	return util.Sorting{
		By:    sortBy,
		Order: orderBy,
	}
}
