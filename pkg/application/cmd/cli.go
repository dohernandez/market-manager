package cmd

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/application/cmd/cli"
	"github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/render"
	"github.com/dohernandez/market-manager/pkg/application/storage"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
)

type (
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

	if cliCtx.String("stock") != "" {
		_, err := bus.ExecuteContext(ctx, &command.UpdateOneStockPrice{Symbol: cliCtx.String("stock")})
		if err != nil {
			return err
		}
	} else if cliCtx.String("wallet") != "" {
		_, err := bus.ExecuteContext(ctx, &command.UpdateWalletStocksPrice{Wallet: cliCtx.String("wallet")})
		if err != nil {
			return err
		}
	} else {
		_, err := bus.ExecuteContext(ctx, &command.UpdateAllStocksPrice{})
		if err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Update finished")

	return nil
}

func (cmd *CLI) UpdateDividend(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	if cliCtx.String("stock") != "" {
		_, err := bus.ExecuteContext(ctx, &command.UpdateOneStockDividend{
			Symbol: cliCtx.String("stock"),
		})
		if err != nil {
			return err
		}
	} else if cliCtx.String("wallet") != "" {
		_, err := bus.ExecuteContext(ctx, &command.UpdateWalletStocksDividend{Wallet: cliCtx.String("wallet")})
		if err != nil {
			return err
		}
	} else {
		_, err := bus.ExecuteContext(ctx, &command.UpdateAllStockDividend{})
		if err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Update finished")

	return nil
}

func (cmd *CLI) ImportStock(cliCtx *cli.Context) error {
	ctxt, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	return cmd_cli.Import(
		ctxt,
		func(_ context.Context, ri cmd_cli.ResourceImport) (interface{}, error) {
			bus := cmd.initCommandBus()

			return bus.ExecuteContext(ctxt, &command.ImportStock{FilePath: ri.FilePath})
		},
		cmd.resourceStorage,
		"stocks",
		cmd.config.Import.StocksPath,
		cliCtx.String("file"),
		"",
	)
}

func (cmd *CLI) ImportTransfer(cliCtx *cli.Context) error {
	ctxt, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	return cmd_cli.Import(
		ctxt,
		func(ctx context.Context, ri cmd_cli.ResourceImport) (interface{}, error) {
			bus := cmd.initCommandBus()

			return bus.ExecuteContext(ctx, &command.ImportTransfer{FilePath: ri.FilePath})
		},
		cmd.resourceStorage,
		"transfers",
		cmd.config.Import.TransfersPath,
		cliCtx.String("file"),
		"",
	)
}

// ImportWallet
func (cmd *CLI) ImportWallet(cliCtx *cli.Context) error {
	ctxt, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	return cmd_cli.Import(
		ctxt,
		func(ctx context.Context, ri cmd_cli.ResourceImport) (interface{}, error) {
			bus := cmd.initCommandBus()

			return bus.ExecuteContext(ctx, &command.ImportWallet{FilePath: ri.FilePath, Name: ri.ResourceName})
		},
		cmd.resourceStorage,
		"wallets",
		cmd.config.Import.WalletsPath,
		cliCtx.String("file"),
		cliCtx.String("wallet"),
	)
}

func (cmd *CLI) ImportOperation(cliCtx *cli.Context) error {
	ctxt, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	return cmd_cli.Import(
		ctxt,
		func(ctx context.Context, ri cmd_cli.ResourceImport) (interface{}, error) {
			bus := cmd.initCommandBus()

			return bus.ExecuteContext(ctx, &command.ImportOperation{FilePath: ri.FilePath, Wallet: ri.ResourceName})
		},
		cmd.resourceStorage,
		"accounts",
		cmd.config.Import.AccountsPath,
		cliCtx.String("file"),
		"",
	)
}

func (cmd *CLI) ImportDividendRetention(cliCtx *cli.Context) error {
	ctxt, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctxt).Fatal("Missing wallet name")
	}

	return cmd_cli.Import(
		ctxt,
		func(ctx context.Context, ri cmd_cli.ResourceImport) (interface{}, error) {
			bus := cmd.initCommandBus()

			return bus.ExecuteContext(ctx, &command.ImportRetention{FilePath: ri.FilePath, Wallet: ri.ResourceName})
		},
		cmd.resourceStorage,
		"retentions",
		cmd.config.Import.RetentionsPath,
		cliCtx.String("file"),
		cliCtx.String("wallet"),
	)
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

// ExportWalletDetails List in csv format or print into screen the wallet details
func (cmd *CLI) ExportWallet(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	bus := cmd.initCommandBus()

	sells := map[string]int{}
	strSells := cliCtx.String("sells")
	if strSells != "" {
		sSells := strings.Split(strSells, ",")

		for _, sSell := range sSells {
			sa := strings.Split(sSell, ":")
			a, _ := strconv.Atoi(sa[1])
			sells[sa[0]] = a
		}
	}

	buys := map[string]int{}
	strBuys := cliCtx.String("buys")
	if strBuys != "" {
		sBuys := strings.Split(strBuys, ",")

		for _, sBuy := range sBuys {
			ba := strings.Split(sBuy, ":")
			a, _ := strconv.Atoi(ba[1])
			buys[ba[0]] = a
		}
	}

	commissions, err := cmd.getCommissionsToApplyStockOperation()
	if (len(sells) > 0 || len(buys) > 0) && err != nil {
		return errors.Wrapf(err, "Can execute buys or sells without commissions")
	}

	var status operation.Status
	switch cliCtx.String("status") {
	case "inactive":
		status = operation.Inactive
	case "all":
		status = operation.All
	default:
		status = operation.Active
	}

	wOutput, err := bus.ExecuteContext(ctx, &command.WalletDetails{
		Wallet:             cliCtx.String("wallet"),
		Sells:              sells,
		Buys:               buys,
		Commissions:        commissions,
		Status:             status,
		IncreaseInvestment: cliCtx.String("transfer"),
	})
	if err != nil {
		return err
	}

	rOutput := render.OutputScreenWalletDetails{
		WalletDetails: wOutput.(render.WalletDetailsOutput),
		Sorting:       cmd.sortingFromCliCtx(cliCtx),
		Precision:     2,
	}

	if cliCtx.String("stock") != "" {
		sls := render.NewScreenWalletStockDetails(cmd.ctx)
		sls.Render(&render.OutputScreenWalletStockDetails{
			OutputScreenWalletDetails: rOutput,
			Stock: cliCtx.String("stock"),
		})

		return nil
	}

	sls := render.NewScreenWalletDetails()
	sls.Render(&rOutput)

	return nil
}

func (cmd *CLI) getCommissionsToApplyStockOperation() (map[string]command.Commission, error) {
	exchanges := cmd.config.Degiro.Exchanges
	commissions := map[string]command.Commission{}

	// NASDAQ
	cCommission, err := json.Marshal(exchanges.NASDAQ)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not marshal config commission")
	}

	var commission command.Commission
	err = json.Unmarshal(cCommission, &commission)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not unmarshal config commission")
	}

	commissions["NASDAQ"] = commission

	// NYSE
	cCommission, err = json.Marshal(exchanges.NYSE)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not marshal config commission")
	}

	commission = command.Commission{}
	err = json.Unmarshal(cCommission, &commission)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not unmarshal config commission")
	}

	commissions["NYSE"] = commission

	return commissions, err
}

// ReloadWallet reload the wallet operation
func (cmd *CLI) ReloadWallet(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	bus := cmd.initCommandBus()

	_, err := bus.ExecuteContext(ctx, &command.ReloadWallet{
		Wallet: cliCtx.String("wallet"),
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed reloading wallet")
	}

	logger.FromContext(ctx).Info("Reloading finished")

	return nil
}

func (cmd *CLI) AddDividend(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	if cliCtx.String("date") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's date")
	}

	if cliCtx.String("stock") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's stock")
	}

	if cliCtx.String("value") == "" {
		logger.FromContext(ctx).Fatal("Missing dividend value")
	}

	bus := cmd.initCommandBus()

	_, err := bus.ExecuteContext(ctx, &command.AddDividendOperation{
		Wallet: cliCtx.String("wallet"),
		Date:   cliCtx.String("date"),
		Stock:  cliCtx.String("stock"),
		Value:  cliCtx.Float64("value"),
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed adding dividend operation to the wallet")
	}

	logger.FromContext(ctx).Info("Adding dividend operation to the wallet finished")

	return nil
}

func (cmd *CLI) AddBuyStock(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	if cliCtx.String("trade") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's trade")
	}

	if cliCtx.String("date") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's date")
	}

	if cliCtx.String("stock") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's stock")
	}

	if cliCtx.String("amount") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's stock amount")
	}

	if cliCtx.String("value") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's value")
	}

	if cliCtx.String("price") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's price")
	}

	if cliCtx.String("price-change") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's price change")
	}

	if cliCtx.String("price-change-commission") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's price change commission")
	}

	if cliCtx.String("commission") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's commission")
	}

	bus := cmd.initCommandBus()

	_, err := bus.ExecuteContext(ctx, &command.AddBuyOperation{
		Trade:                 cliCtx.String("trade"),
		Wallet:                cliCtx.String("wallet"),
		Date:                  cliCtx.String("date"),
		Stock:                 cliCtx.String("stock"),
		Value:                 cliCtx.Float64("value"),
		Price:                 cliCtx.Float64("price"),
		PriceChange:           cliCtx.Float64("price-change"),
		PriceChangeCommission: cliCtx.Float64("price-change-commission"),
		Commission:            cliCtx.Float64("commission"),
		Amount:                cliCtx.Int("amount"),
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed adding buy operation to the wallet")
	}

	logger.FromContext(ctx).Info("Adding buy operation to the wallet finished")

	return nil
}

func (cmd *CLI) AddSellStock(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	if cliCtx.String("trade") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's trade")
	}

	if cliCtx.String("date") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's date")
	}

	if cliCtx.String("stock") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's stock")
	}

	if cliCtx.String("amount") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's stock amount")
	}

	if cliCtx.String("value") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's value")
	}

	if cliCtx.String("price") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's price")
	}

	if cliCtx.String("price-change") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's price change")
	}

	if cliCtx.String("price-change-commission") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's price change commission")
	}

	if cliCtx.String("commission") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's commission")
	}

	bus := cmd.initCommandBus()

	_, err := bus.ExecuteContext(ctx, &command.AddSellOperation{
		Trade:                 cliCtx.String("trade"),
		Wallet:                cliCtx.String("wallet"),
		Date:                  cliCtx.String("date"),
		Stock:                 cliCtx.String("stock"),
		Value:                 cliCtx.Float64("value"),
		Price:                 cliCtx.Float64("price"),
		PriceChange:           cliCtx.Float64("price-change"),
		PriceChangeCommission: cliCtx.Float64("price-change-commission"),
		Commission:            cliCtx.Float64("commission"),
		Amount:                cliCtx.Int("amount"),
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed sell dividend operation to the wallet")
	}

	logger.FromContext(ctx).Info("Adding sell operation to the wallet finished")

	return nil
}

func (cmd *CLI) AddInterest(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	if cliCtx.String("date") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's date")
	}

	if cliCtx.String("value") == "" {
		logger.FromContext(ctx).Fatal("Missing interest value")
	}

	bus := cmd.initCommandBus()

	_, err := bus.ExecuteContext(ctx, &command.AddInterestOperation{
		Wallet: cliCtx.String("wallet"),
		Date:   cliCtx.String("date"),
		Value:  cliCtx.Float64("value"),
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed adding interest operation to the wallet")
	}

	logger.FromContext(ctx).Info("Adding interest operation to the wallet finished")

	return nil
}

func (cmd *CLI) ExportSnapshotWallet(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet name")
	}

	if cliCtx.String("date") == "" {
		logger.FromContext(ctx).Fatal("Missing operation's date")
	}

	var excludes []string
	if cliCtx.String("exclude") != "" {
		excludes = strings.Split(cliCtx.String("exclude"), ",")
	}

	bus := cmd.initCommandBus()

	wOutput, err := bus.ExecuteContext(ctx, &command.WalletDateDetails{
		Wallet:        cliCtx.String("wallet"),
		Date:          cliCtx.String("date"),
		TransferPath:  cmd.config.Import.TransfersPath,
		OperationPath: cmd.config.Import.AccountsPath,
		Excludes:      excludes,
	})
	if err != nil {
		return err
	}

	rOutput := render.OutputScreenWalletDetails{
		WalletDetails: wOutput.(render.WalletDetailsOutput),
		Sorting:       cmd.sortingFromCliCtx(cliCtx),
		Precision:     2,
	}

	if cliCtx.String("stock") != "" {
		sls := render.NewScreenWalletStockDetails(cmd.ctx)
		sls.Render(&render.OutputScreenWalletStockDetails{
			OutputScreenWalletDetails: rOutput,
			Stock: cliCtx.String("stock"),
		})

		return nil
	}

	sls := render.NewWalletDateDetails(cmd.ctx)
	sls.Render(&rOutput)

	return nil
}

// AddStock adds stock. Scraped the rest of information of the stock from Yahoo/MarketChameleon
func (cmd *CLI) AddStock(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("stock") == "" {
		logger.FromContext(ctx).Fatal("Missing stock symbol")
	}

	if cliCtx.String("exchange") == "" {
		logger.FromContext(ctx).Fatal("Missing exchange symbol")
	}

	bus := cmd.initCommandBus()

	_, err := bus.ExecuteContext(ctx, &command.AddStock{
		Symbol:   cliCtx.String("stock"),
		Exchange: cliCtx.String("exchange"),
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed adding stock")
	}

	logger.FromContext(ctx).Info("Add stock finished")

	return nil
}

// AddDividendRetention adds dividend retention.
func (cmd *CLI) AddDividendRetention(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	if cliCtx.String("wallet") == "" {
		logger.FromContext(ctx).Fatal("Missing wallet symbol")
	}

	if cliCtx.String("stock") == "" {
		logger.FromContext(ctx).Fatal("Missing stock symbol")
	}

	if cliCtx.String("retention") == "" {
		logger.FromContext(ctx).Fatal("Missing retention value")
	}

	bus := cmd.initCommandBus()

	_, err := bus.ExecuteContext(ctx, &command.AddDividendRetention{
		Stock:     cliCtx.String("stock"),
		Retention: cliCtx.String("retention"),
		Wallet:    cliCtx.String("wallet"),
	})
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed adding stock")
	}

	logger.FromContext(ctx).Info("Add dividend retention finished")

	return nil
}
