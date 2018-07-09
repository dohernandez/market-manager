package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type updateWalletCapital struct {
	walletFinder    wallet.Finder
	walletPersister wallet.Persister

	ccClient *cc.Client
}

func NewUpdateWalletCapital(walletFinder wallet.Finder, walletPersister wallet.Persister, ccClient *cc.Client) *updateWalletCapital {
	return &updateWalletCapital{
		walletFinder:    walletFinder,
		walletPersister: walletPersister,
		ccClient:        ccClient,
	}
}

func (l *updateWalletCapital) OnEvent(ctx context.Context, event cbus.Event) {
	stks := event.Result.([]*stock.Stock)

	cEURUSD, err := l.ccClient.Converter.Get()
	if err != nil {
		logger.FromContext(ctx).Error("An error happen while getting converter EURUSD: error [%s]")

		return
	}

	for _, stk := range stks {
		ws, err := l.walletFinder.FindWithItemByStock(stk)
		if err != nil {
			if err != mm.ErrNotFound {
				logger.FromContext(ctx).Errorf(
					"An error happen while finding stocks item in wallet: stock [%s] -> error [%s]",
					stk.Symbol,
					err,
				)

				return
			}

			continue
		}

		for _, w := range ws {
			w.Items[stk.ID].CapitalRate = cEURUSD.EURUSD

			capital := w.Items[stk.ID].Capital()
			w.Capital = capital
		}

		err = l.walletPersister.UpdateAllItemsCapital(ws)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while updating all stock item capital in wallet: error [%s]",
				err,
			)

			return
		}
	}
}
