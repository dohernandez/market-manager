package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
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
	var stks []*stock.Stock
	stks, ok := event.Result.([]*stock.Stock)
	if !ok {
		ops, ok := event.Result.([]*operation.Operation)
		if !ok {
			logger.FromContext(ctx).Warn("updateWalletCapital: Result instance not supported")

			return
		}

		astks := map[uuid.UUID]*stock.Stock{}
		for _, o := range ops {
			if _, ok := astks[o.Stock.ID]; !ok {
				stks = append(stks, o.Stock)

				astks[o.Stock.ID] = o.Stock
			}
		}
	}

	currencyConverter, err := l.ccClient.Converter.Get()
	if err != nil {
		logger.FromContext(ctx).Error("An error happen while getting currency to euro converter: error [%s]")

		return
	}

	//panic(fmt.Sprintf("%+v", currencyConverter))

	capitalRate := wallet.CapitalRate{
		EURUSD: currencyConverter.EURUSD,
		EURCAD: currencyConverter.EURCAD,
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
			w.SetCapitalRate(capitalRate)
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

	logger.FromContext(ctx).Debug("Updated wallet capital")
}
