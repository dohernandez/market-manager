package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type persisterWallet struct {
	walletPersister wallet.Persister
}

func NewPersisterWallet(walletPersister wallet.Persister) *persisterWallet {
	return &persisterWallet{
		walletPersister: walletPersister,
	}
}

func (l *persisterWallet) OnEvent(ctx context.Context, event cbus.Event) {
	ws := event.Result.([]*wallet.Wallet)

	err := l.walletPersister.PersistAll(ws)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting wallets -> error [%s]",
			err,
		)

		return
	}
}
