package handler

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type importRetention struct {
	stockFinder     stock.Finder
	walletFinder    wallet.Finder
	walletPersister wallet.Persister
}

func NewImportRetention(
	stockFinder stock.Finder,
	walletFinder wallet.Finder,
	walletPersister wallet.Persister,
) *importRetention {
	return &importRetention{
		stockFinder:     stockFinder,
		walletFinder:    walletFinder,
		walletPersister: walletPersister,
	}
}

func (h *importRetention) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportRetention).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	wName := command.(*appCommand.ImportRetention).Wallet
	if wName == "" {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet -> error [%s]",
			err,
		)

		return nil, errors.New("missing wallet name")
	}

	w, err := h.LoadWalletWithActiveWalletItems(wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet name [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		for _, i := range w.Items {
			if i.Stock.Name != line[0] {
				continue
			}

			i.DividendRetention = mm.ValueDollarFromString(line[1])
		}
	}

	err = h.walletPersister.UpdateRetentions(w)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting operation -> error [%s]",
			err,
		)

		return nil, err
	}

	return w, nil
}

func (h *importRetention) LoadWalletWithActiveWalletItems(name string) (*wallet.Wallet, error) {
	w, err := h.walletFinder.FindByName(name)
	if err != nil {
		return nil, err
	}

	if err = h.walletFinder.LoadActiveItems(w); err != nil {
		return nil, err
	}

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := h.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return nil, err
		}

		// I like to keep the address but change the content to keep,
		// trade and item pointing to the same stock
		*i.Stock = *stk
	}

	return w, err
}

// parseDateString - parse a potentially partial date string to Time
func (h *importRetention) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2.1.2006", dt)

	return t
}
