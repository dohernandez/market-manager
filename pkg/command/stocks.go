package command

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli"

	"os"
	"sort"
	"text/tabwriter"

	"strings"

	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// StocksCommand ...
type StocksCommand struct {
	*BaseCommand
}

// NewStocksCommand constructs StocksCommand
func NewStocksCommand(baseCommand *BaseCommand) *StocksCommand {
	return &StocksCommand{
		BaseCommand: baseCommand,
	}
}

// Run runs the application stock update
func (cmd *StocksCommand) Run(cliCtx *cli.Context) error {
	return cmd.Price(cliCtx)
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// START Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// ByName implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded wallet items value.
type Stocks []*stock.Stock

func (s Stocks) Len() int      { return len(s) }
func (s Stocks) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type StocksByName struct {
	Stocks
}

func (s StocksByName) Less(i, j int) bool {
	return s.Stocks[i].Name < s.Stocks[j].Name
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// END Stocks Sort
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// List in csv format the wallet items from a wallet
func (cmd *StocksCommand) Stocks(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Database connection
	logger.FromContext(ctx).Info("Initializing database connection")
	db, err := cmd.initDatabaseConnection()
	if err != nil {
		logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	}

	c := cmd.Container(db)

	var stks []*stock.Stock
	if cliCtx.String("exchange") == "" {
		stks, err = c.PurchaseServiceInstance().Stocks()
	} else {
		exchanges := strings.Split(cliCtx.String("exchange"), ",")

		stks, err = c.PurchaseServiceInstance().StocksByExchanges(exchanges)
	}
	if err != nil {
		return err
	}

	tabw := cmd.formatStocksToScreen(stks)
	tabw.Flush()

	return nil
}

// formatStocksToScreen - convert Items structure to csv string
func (cmd *StocksCommand) formatStocksToScreen(stks []*stock.Stock) *tabwriter.Writer {
	precision := 2
	sortStks := make([]*stock.Stock, 0, len(stks))

	for _, stk := range stks {
		sortStks = append(sortStks, stk)
	}

	sort.Sort(StocksByName{sortStks})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	fmt.Fprintln(w, "#\t Stock\t Market\t Symbol\t Value\t Dividend Yield\t Change\t Last Price Update\t")
	for i, stk := range sortStks {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %.*f\t %.*f\t %.*f\t %s\t",
			i+1,
			stk.Name,
			stk.Exchange.Symbol,
			stk.Symbol,
			precision,
			stk.Value.Amount,
			precision,
			stk.DividendYield,
			precision,
			stk.Change.Amount,
			stk.LastPriceUpdate.Format(time.RFC822),
		)
		fmt.Fprintln(w, str)
	}

	return w
}
