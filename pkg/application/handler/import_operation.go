package handler

import (
	"context"
	"io"

	"github.com/gogolfing/cbus"

	"github.com/satori/go.uuid"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type importOperation struct {
	stockFinder stock.Finder
}

func NewImportOperation(
	stockFinder stock.Finder,
) *importOperation {
	return &importOperation{
		stockFinder: stockFinder,
	}
}

func (h *importOperation) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportOperation).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	var os []*operation.Operation
	trades := map[uuid.UUID]string{}

	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		o, err := createOperationFromLine(line, h.stockFinder)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while createOperationFromLine %s -> error [%s]",
				line,
				err,
			)

			return nil, err
		}

		os = append(os, o)

		if line[0] != "" {
			trades[o.ID] = line[0]
		}
	}

	command.(*appCommand.ImportOperation).Trades = trades

	return os, nil
}
