package handler

import (
	"context"
	"errors"
	"io"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type importRetention struct {
	stockFinder  stock.Finder
	walletFinder wallet.Finder
}

func NewImportRetention(
	stockFinder stock.Finder,
	walletFinder wallet.Finder,
) *importRetention {
	return &importRetention{
		stockFinder:  stockFinder,
		walletFinder: walletFinder,
	}
}

func (h *importRetention) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportRetention).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	wName := command.(*appCommand.ImportRetention).Wallet
	if wName == "" {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet -> error [%s]",
			err,
		)

		return nil, errors.New("missing wallet name")
	}

	w, err := loadWalletWithActiveWalletItems(h.walletFinder, h.stockFinder, wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet name [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		for _, i := range w.Items {
			if i.Stock.Name != line[0] {
				continue
			}

			i.DividendRetention = mm.ValueDollarFromString(line[1])
		}
	}

	return w, nil
}
