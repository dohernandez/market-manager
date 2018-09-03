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
	ops, ok := event.Result.([]*operation.Operation)
	if !ok {
		logger.FromContext(ctx).Warn("addWalletOperation: Result instance not supported")

		return
	}

	var wName string
	trades := map[uuid.UUID]string{}

	switch cmd := event.Command.(type) {
	case *appCommand.AddDividendOperation:
		wName = cmd.Wallet
	case *appCommand.AddBuyOperation:
		wName = cmd.Wallet
		trades[ops[0].ID] = cmd.Trade
	case *appCommand.AddSellOperation:
		wName = cmd.Wallet
		trades[ops[0].ID] = cmd.Trade
	case *appCommand.ImportOperation:
		wName = cmd.Wallet
		trades = cmd.Trades
	default:
		logger.FromContext(ctx).Error(
			"addWalletOperation: Operation action not supported",
		)

		return
	}

	if wName == "" {
		logger.FromContext(ctx).Error(
			"An error happen while loading wallet -> error [%s]",
			wName,
		)

		return
	}

	w, err := l.loadWalletWithActiveWalletItemsAndActiveWalletTrades(wName)
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

func (l *addWalletOperation) loadWalletWithActiveWalletItemsAndActiveWalletTrades(name string) (*wallet.Wallet, error) {
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
