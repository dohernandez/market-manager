package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type persisterOperation struct {
	walletFinder    wallet.Finder
	stockFinder     stock.Finder
	walletPersister wallet.Persister
	ccClient        *cc.Client
}

func NewPersisterOperation(walletFinder wallet.Finder, stockFinder stock.Finder, walletPersister wallet.Persister, ccClient *cc.Client) *persisterOperation {
	return &persisterOperation{
		walletFinder:    walletFinder,
		stockFinder:     stockFinder,
		walletPersister: walletPersister,
		ccClient:        ccClient,
	}
}

func (l *persisterOperation) OnEvent(ctx context.Context, event cbus.Event) {
	os := event.Result.([]*operation.Operation)

	cEURUSD, err := l.ccClient.Converter.Get()
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while getting EUR to USD change -> error [%s]",
			err,
		)

		return
	}

	wName := event.Command.(*appCommand.ImportOperation).Wallet
	if wName == "" {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet -> error [%s]",
			err,
		)

		return
	}

	w, err := l.walletFinder.FindByName(wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet name [%s] -> error [%s]",
			wName,
			err,
		)

		return
	}

	err = l.LoadActiveWalletItems(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet active items [%s] -> error [%s]",
			wName,
			err,
		)

		return
	}

	for _, o := range os {
		w.AddOperation(o)
	}

	for _, wItem := range w.Items {
		wItem.CapitalRate = cEURUSD.EURUSD

		capital := wItem.Capital()
		w.Capital = capital
	}

	err = l.walletPersister.PersistOperations(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting operation -> error [%s]",
			err,
		)

		return
	}
}

func (l *persisterOperation) LoadActiveWalletItems(w *wallet.Wallet) error {
	err := l.walletFinder.LoadActiveItems(w)
	if err != nil {
		return err
	}

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := l.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return err
		}

		i.Stock = stk
	}

	return nil
}
