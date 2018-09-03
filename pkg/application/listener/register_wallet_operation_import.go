package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"regexp"

	"fmt"

	"strconv"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
)

type registerWalletOperationImport struct {
	resourceStorage util.ResourceStorage
	importPath      string
}

const linePerFile = 10

func NewRegisterWalletOperationImport(resourceStorage util.ResourceStorage, importPath string) *registerWalletOperationImport {
	return &registerWalletOperationImport{
		resourceStorage: resourceStorage,
		importPath:      importPath,
	}
}

func (l *registerWalletOperationImport) OnEvent(ctx context.Context, event cbus.Event) {
	var wName string

	aDividend, ok := event.Command.(*appCommand.AddDividend)
	if ok {
		wName = aDividend.Wallet
	} else {
		logger.FromContext(ctx).Warn("Result instance not supported")

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
		fileNumber := l.geResourceNumberFromFilePath(r.FileName)

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

	ops, ok := event.Result.([]*operation.Operation)
	if !ok {
		logger.FromContext(ctx).Warn("Result instance not supported")

		return
	}

	for _, o := range ops {
		switch o.Action {
		case operation.Dividend:
			v := fmt.Sprintf("%.2f", o.Value.Amount)

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
		default:
			logger.FromContext(ctx).Warn("Operation action not supported")

			return
		}
	}

	err = wf.WriteAllLines(lines)
	if err != nil {
		logger.FromContext(ctx).Warn(err)

		return
	}
}

func (l *registerWalletOperationImport) geResourceNumberFromFilePath(fileName string) int {
	reg := regexp.MustCompile(`(^[0-9]{2})+_+(.*)`)
	res := reg.ReplaceAllString(fileName, "${1}")

	n, _ := strconv.Atoi(res)

	return n
}
