package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type saveDividendRetention struct {
	walletPersister wallet.Persister
}

func NewSaveDividendRetention(walletPersister wallet.Persister) *saveDividendRetention {
	return &saveDividendRetention{
		walletPersister: walletPersister,
	}
}

func (l *saveDividendRetention) OnEvent(ctx context.Context, event cbus.Event) {
	w, ok := event.Result.(*wallet.Wallet)
	if !ok {
		logger.FromContext(ctx).Warn("saveDividendRetention: Result instance not supported")

		return
	}

	err := l.walletPersister.UpdateRetentions(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting operation -> error [%s]",
			err,
		)
	}
}
