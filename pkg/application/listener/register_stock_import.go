package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type registerStockImport struct {
	resourceStorage util.ResourceStorage
	importPath      string
}

func NewRegisterStockImport(resourceStorage util.ResourceStorage, importPath string) *registerStockImport {
	return &registerStockImport{
		resourceStorage: resourceStorage,
		importPath:      importPath,
	}
}

func (l *registerStockImport) OnEvent(ctx context.Context, event cbus.Event) {
	ss, ok := event.Result.([]*stock.Stock)
	if !ok {
		logger.FromContext(ctx).Warn("registerStockImport: Result instance not supported")

		return
	}

	wf, err := getCsvWriterFromResourceImport(l.resourceStorage, "stocks", l.importPath, "stocks")
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while getting csv writer from import for stocks -> error [%s]",
			err,
		)
	}

	wf.Open()
	defer wf.Close()
	defer wf.Flush()

	var lines [][]string
	for _, s := range ss {
		line := []string{
			s.Name,
			s.Exchange.Symbol,
			s.Symbol,
			s.Type.Name,
			s.Sector.Name,
			s.Industry.Name,
		}

		lines = append(lines, line)
	}

	err = wf.WriteAllLines(lines)
	if err != nil {
		logger.FromContext(ctx).Warn(err)

		return
	}

	logger.FromContext(ctx).Debug("Registered stocks into file")
}
