package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/satori/go.uuid"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/storage"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type reloadWallet struct {
	walletFinder wallet.Finder
	walletReload *storage.WalletReload
}

func NewReloadWallet(walletFinder wallet.Finder, walletReload *storage.WalletReload) *reloadWallet {
	return &reloadWallet{
		walletFinder: walletFinder,
		walletReload: walletReload,
	}
}

func (h *reloadWallet) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	wName := command.(*appCommand.ReloadWallet).Wallet

	w, err := h.walletFinder.FindByName(wName)
	if err != nil {
		return nil, err
	}

	err = h.walletReload.Reload(w)
	if err != nil {
		return nil, err
	}

	w.Items = map[uuid.UUID]*wallet.Item{}
	w.Invested = mm.Value{}
	w.Capital = mm.Value{}
	w.Funds = mm.Value{}
	w.Dividend = mm.Value{}
	w.Commission = mm.Value{}
	w.Connection = mm.Value{}
	w.Interest = mm.Value{}
	w.Operations = make([]*operation.Operation, 0)

	return w, nil
}
