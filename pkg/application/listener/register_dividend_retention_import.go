package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"time"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type registerDividendRetentionImport struct {
	resourceStorage util.ResourceStorage
	importPath      string
}

func NewRegisterDividendRetentionImport(resourceStorage util.ResourceStorage, importPath string) *registerDividendRetentionImport {
	return &registerDividendRetentionImport{
		resourceStorage: resourceStorage,
		importPath:      importPath,
	}
}

func (l *registerDividendRetentionImport) OnEvent(ctx context.Context, event cbus.Event) {
	addDividendRetentionCommand, ok := event.Command.(*appCommand.AddDividendRetention)
	if !ok {
		logger.FromContext(ctx).Warn("registerDividendRetentionImport: command instance not supported")

		return
	}

	oWallet, ok := event.Result.(*wallet.Wallet)
	if !ok {
		logger.FromContext(ctx).Warn("registerDividendRetentionImport: Result instance not supported")

		return
	}

	wf, err := getCsvWriterFromResourceImport(l.resourceStorage, "retentions", l.importPath, oWallet.Name)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while getting csv writer from import for retentions -> error [%s]",
			err,
		)
	}

	wf.Open()
	defer wf.Close()
	defer wf.Flush()

	var lines [][]string
	now := time.Now()

	for _, i := range oWallet.Items {
		if i.Stock.Symbol != addDividendRetentionCommand.Stock {
			continue
		}

		lines = append(lines, []string{
			i.Stock.Name,
			addDividendRetentionCommand.Retention,
			now.Format("2.1.2006"),
		})
	}

	err = wf.WriteAllLines(lines)
	if err != nil {
		logger.FromContext(ctx).Warn(err)

		return
	}

	logger.FromContext(ctx).Debug("Registered dividend retention into file")
}
