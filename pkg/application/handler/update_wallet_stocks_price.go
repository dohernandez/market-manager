package handler

import (
	"context"
	"errors"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type updateWalletStocksPrice struct {
	walletFinder wallet.Finder
	stockFinder  stock.Finder
}

func NewUpdateWalletStocksPrice(walletFinder wallet.Finder, stockFinder stock.Finder) *updateWalletStocksPrice {
	return &updateWalletStocksPrice{
		walletFinder: walletFinder,
		stockFinder:  stockFinder,
	}
}

func (h *updateWalletStocksPrice) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	wName := command.(*appCommand.UpdateWalletStocksPrice).Wallet

	if wName == "" {
		logger.FromContext(ctx).Error("An error happen while loading wallet -> error [wallet can not be empty]")

		return nil, errors.New("missing wallet name")
	}

	w, err := h.walletFinder.FindByName(wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	err = h.walletFinder.LoadActiveItems(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading stocks wallet [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	var stks []*stock.Stock
	for _, i := range w.Items {
		stk, err := h.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while finding stock wallet [%s] ID [%s] -> error [%s]",
				wName,
				i.Stock.ID,
				err,
			)

			return nil, err
		}

		stks = append(stks, stk)
	}

	return stks, nil
}
