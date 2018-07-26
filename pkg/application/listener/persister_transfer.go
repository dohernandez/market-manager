package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type persisterTransfer struct {
	transferPersister transfer.Persister
	walletFinder      wallet.Finder
	walletPersister   wallet.Persister
}

func NewPersisterTransfer(transferPersister transfer.Persister, walletFinder wallet.Finder, walletPersister wallet.Persister) *persisterTransfer {
	return &persisterTransfer{
		transferPersister: transferPersister,
		walletFinder:      walletFinder,
		walletPersister:   walletPersister,
	}
}

func (l *persisterTransfer) OnEvent(ctx context.Context, event cbus.Event) {
	ts := event.Result.([]*transfer.Transfer)

	err := l.transferPersister.PersistAll(ts)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting stocks -> error [%s]",
			err,
		)

		return
	}

	var ws []*wallet.Wallet

	wsf := map[uuid.UUID]*wallet.Wallet{}
	wst := map[uuid.UUID]*wallet.Wallet{}

	for _, t := range ts {
		var err error

		w, ok := wsf[t.From.ID]
		if !ok {
			w, err = l.walletFinder.FindByBankAccount(t.From)
			if err != nil {
				if err != mm.ErrNotFound {
					logger.FromContext(ctx).Errorf(
						"An error happen while finding wallet by bank account from [%s] -> error [%s]",
						t.From.AccountNo,
						err,
					)
					return
				}

				w, ok = wst[t.To.ID]
				if !ok {
					w, err = l.walletFinder.FindByBankAccount(t.To)
					if err != nil {
						logger.FromContext(ctx).Errorf(
							"An error happen while finding wallet by bank account to [%s] -> error [%s]",
							t.To.AccountNo,
							err,
						)
						return
					}

					wst[t.To.ID] = w
					ws = append(ws, w)
				}

				w.IncreaseInvestment(t.Amount)
				continue
			}

			wsf[t.From.ID] = w
			ws = append(ws, w)
		}

		w.DecreaseInvestment(t.Amount)
	}

	l.walletPersister.UpdateAllAccounting(ws)
}
