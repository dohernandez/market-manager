package listener

import (
	"context"
	"strconv"

	"github.com/gogolfing/cbus"
	"github.com/satori/go.uuid"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type addWalletOperation struct {
	stockFinder     stock.Finder
	walletFinder    wallet.Finder
	walletPersister wallet.Persister

	ccClient *cc.Client
}

func NewAddWalletOperation(stockFinder stock.Finder, walletFinder wallet.Finder, walletPersister wallet.Persister, ccClient *cc.Client) *addWalletOperation {
	return &addWalletOperation{
		stockFinder:     stockFinder,
		walletFinder:    walletFinder,
		walletPersister: walletPersister,
		ccClient:        ccClient,
	}
}

func (l *addWalletOperation) OnEvent(ctx context.Context, event cbus.Event) {
	var wName string
	trades := map[uuid.UUID]string{}

	iOperation, ok := event.Command.(*appCommand.ImportOperation)
	if ok {
		wName = iOperation.Wallet
		trades = iOperation.Trades
	} else {
		aDividend, ok := event.Command.(*appCommand.AddDividend)
		if !ok {
			logger.FromContext(ctx).Warn("Result instance not supported")

			return
		}

		wName = aDividend.Wallet
	}

	if wName == "" {
		logger.FromContext(ctx).Error(
			"An error happen while loading wallet -> error [%s]",
			wName,
		)

		return
	}

	ops, ok := event.Result.([]*operation.Operation)
	if !ok {
		logger.FromContext(ctx).Warn("Result instance not supported")

		return
	}

	w, err := l.LoadWalletWithActiveWalletItemsAndActiveWalletTrades(wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet name [%s] -> error [%s]",
			wName,
			err,
		)

		return
	}

	for _, o := range ops {
		w.AddOperation(o)

		nTrade, ok := trades[o.ID]
		if ok {
			n, _ := strconv.Atoi(nTrade)
			w.AddTrade(n, o)
		} else if o.Action == operation.Dividend {
			w.AddTrade(0, o)
		}
	}

	err = l.walletPersister.PersistOperations(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting operation -> error [%s]",
			err,
		)
	}
}

func (l *addWalletOperation) LoadWalletWithActiveWalletItemsAndActiveWalletTrades(name string) (*wallet.Wallet, error) {
	w, err := l.walletFinder.FindByName(name)
	if err != nil {
		return nil, err
	}

	if err = l.walletFinder.LoadActiveItems(w); err != nil {
		return nil, err
	}

	if err = l.walletFinder.LoadActiveTrades(w); err != nil {
		return nil, err
	}

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := l.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return nil, err
		}

		// I like to keep the address but change the content to keep,
		// trade and item pointing to the same stock
		*i.Stock = *stk
	}

	cEURUSD, err := l.ccClient.Converter.Get()
	if err != nil {
		return nil, err
	}

	w.SetCapitalRate(cEURUSD.EURUSD)

	return w, err
}
