package handler

import (
	"context"
	"errors"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type addDividendRetention struct {
	stockFinder  stock.Finder
	walletFinder wallet.Finder
}

func NewAddDividendRetention(
	stockFinder stock.Finder,
	walletFinder wallet.Finder,
) *addDividendRetention {
	return &addDividendRetention{
		stockFinder:  stockFinder,
		walletFinder: walletFinder,
	}
}

func (h *addDividendRetention) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	addDividendRetentionCommand := command.(*appCommand.AddDividendRetention)

	wName := addDividendRetentionCommand.Wallet
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

	for _, i := range w.Items {
		if i.Stock.Symbol != addDividendRetentionCommand.Stock {
			continue
		}

		i.DividendRetention = mm.ValueDollarFromString(addDividendRetentionCommand.Retention)
	}

	return w, nil
}
