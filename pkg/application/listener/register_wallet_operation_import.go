package listener

import (
	"context"
	"fmt"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
)

type registerWalletOperationImport struct {
	resourceStorage util.ResourceStorage
	importPath      string
}

func NewRegisterWalletOperationImport(resourceStorage util.ResourceStorage, importPath string) *registerWalletOperationImport {
	return &registerWalletOperationImport{
		resourceStorage: resourceStorage,
		importPath:      importPath,
	}
}

func (l *registerWalletOperationImport) OnEvent(ctx context.Context, event cbus.Event) {
	ops, ok := event.Result.([]*operation.Operation)
	if !ok {
		logger.FromContext(ctx).Warn("registerWalletOperationImport: Result instance not supported")

		return
	}

	var wName, trade string

	switch cmd := event.Command.(type) {
	case *appCommand.AddDividendOperation:
		wName = cmd.Wallet
	case *appCommand.AddBuyOperation:
		wName = cmd.Wallet
		trade = cmd.Trade
	case *appCommand.AddSellOperation:
		wName = cmd.Wallet
		trade = cmd.Trade
	case *appCommand.AddInterestOperation:
		wName = cmd.Wallet
	default:
		logger.FromContext(ctx).Error(
			"registerWalletOperationImport: Operation action not supported",
		)

		return
	}

	if wName == "" {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet [%s]",
			wName,
		)

		return
	}

	wf, err := getCsvWriterFromResourceImport(l.resourceStorage, "accounts", l.importPath, wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while getting csv writer from import for wallet [%s] -> error [%s]",
			wName,
			err,
		)
	}

	wf.Open()
	defer wf.Close()
	defer wf.Flush()

	var lines [][]string
	for _, o := range ops {
		var (
			action                operation.Action
			stockName             string
			price                 string
			amount                string
			priceChange           string
			priceChangeCommission string
			commission            string
		)
		v := fmt.Sprintf("%.2f", o.Value.Amount)

		switch o.Action {
		case operation.Dividend:
			action = "Dividendo"
			stockName = o.Stock.Name
			price = v
		case operation.Buy, operation.Sell:
			action = "Compra"
			if o.Action == operation.Sell {
				action = "Venta"
			}

			stockName = o.Stock.Name
			amount = fmt.Sprintf("%d", o.Amount)
			price = fmt.Sprintf("%.2f", o.Price.Amount)
			priceChange = fmt.Sprintf("%.2f", o.PriceChange.Amount)
			priceChangeCommission = fmt.Sprintf("%.2f", o.PriceChangeCommission.Amount)
			commission = fmt.Sprintf("%.2f", o.Commission.Amount)
		case operation.Interest:
			action = "Interés"
			price = v
		default:
			logger.FromContext(ctx).Warn("registerWalletOperationImport: Operation action not supported")

			return
		}

		line := []string{
			trade,
			o.Date.Format("2/1/2006"),
			stockName,
			string(action),
			amount,
			price,
			priceChange,
			priceChangeCommission,
			v,
			commission,
		}

		lines = append(lines, line)
	}

	err = wf.WriteAllLines(lines)
	if err != nil {
		logger.FromContext(ctx).Warn(err)

		return
	}

	logger.FromContext(ctx).Debug("Registered operation wallet into file")
}
