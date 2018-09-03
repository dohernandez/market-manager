package listener

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gogolfing/cbus"

	"regexp"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
)

type registerWalletOperationImport struct {
	resourceStorage util.ResourceStorage
	importPath      string
}

const linePerFile = 30

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

	r, err := l.resourceStorage.FindLastByResourceAndWallet("accounts", wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading latest account resource import from wallet [%s] -> error [%s]",
			wName,
			err,
		)

		return
	}

	var lines [][]string

	filePath := fmt.Sprintf("%s/%s", l.importPath, r.FileName)
	rf := util.NewCsvReader(filePath)

	rf.Open()
	defer rf.Close()

	lines, err = rf.ReadAllLines()
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading previous lines from resource [%s] -> error [%s]",
			filePath,
			err,
		)

		return
	}

	nLines := len(lines)
	if err != nil || nLines >= linePerFile {
		fileNumber := l.getResourceNumberFromFilePath(r.FileName)

		fName := fmt.Sprintf("%d_%s.csv", fileNumber+1, wName)
		if fileNumber < 10 {
			fName = fmt.Sprintf("0%d_%s.csv", fileNumber+1, wName)
		}

		filePath = fmt.Sprintf("%s/%s", l.importPath, fName)

		defer func() {
			ir := util.NewResource("accounts", fName)

			err := l.resourceStorage.Persist(ir)
			if err != nil {
				panic(err)
			}
		}()
	}

	wf := util.NewCsvWriter(filePath)

	wf.Open()
	defer wf.Close()
	defer wf.Flush()

	for _, o := range ops {
		v := fmt.Sprintf("%.2f", o.Value.Amount)

		switch o.Action {
		case operation.Dividend:
			lines = append(lines, []string{
				"",
				o.Date.Format("2/1/2006"),
				o.Stock.Name,
				"Dividendo",
				"",
				v,
				"",
				"",
				v,
				"",
			})
		case operation.Buy, operation.Sell:
			action := "Compra"
			if o.Action == operation.Sell {
				action = "Venta"
			}

			a := fmt.Sprintf("%d", o.Amount)
			p := fmt.Sprintf("%.2f", o.Price.Amount)
			pc := fmt.Sprintf("%.2f", o.PriceChange.Amount)
			pcc := fmt.Sprintf("%.2f", o.PriceChangeCommission.Amount)
			c := fmt.Sprintf("%.2f", o.Commission.Amount)

			lines = append(lines, []string{
				trade,
				o.Date.Format("2/1/2006"),
				o.Stock.Name,
				action,
				a,
				p,
				pc,
				pcc,
				v,
				c,
			})
		default:
			logger.FromContext(ctx).Warn("registerWalletOperationImport: Operation action not supported")

			return
		}
	}

	err = wf.WriteAllLines(lines)
	if err != nil {
		logger.FromContext(ctx).Warn(err)

		return
	}
}

func (l *registerWalletOperationImport) getResourceNumberFromFilePath(fileName string) int {
	reg := regexp.MustCompile(`(^[0-9]{2})+_+(.*)`)
	res := reg.ReplaceAllString(fileName, "${1}")

	n, _ := strconv.Atoi(res)

	return n
}
