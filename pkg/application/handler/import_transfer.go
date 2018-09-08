package handler

import (
	"context"
	"io"

	"github.com/gogolfing/cbus"
	"github.com/satori/go.uuid"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type importTransfer struct {
	bankAccountFinder bank.Finder
	transferPersister transfer.Persister
	walletFinder      wallet.Finder
	walletPersister   wallet.Persister
}

func NewImportTransfer(
	bankAccountFinder bank.Finder,
	transferPersister transfer.Persister,
	walletFinder wallet.Finder,
	walletPersister wallet.Persister,
) *importTransfer {
	return &importTransfer{
		bankAccountFinder: bankAccountFinder,
		transferPersister: transferPersister,
		walletFinder:      walletFinder,
		walletPersister:   walletPersister,
	}
}

func (h *importTransfer) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportTransfer).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	var (
		ts []*transfer.Transfer
		ws []*wallet.Wallet
	)

	wsf := map[uuid.UUID]*wallet.Wallet{}
	wst := map[uuid.UUID]*wallet.Wallet{}

	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		t, err := createTransferFromLine(line, h.bankAccountFinder)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while creating transfer -> error [%s]",
				err,
			)
			return nil, err
		}

		ts = append(ts, t)

		w, ok := wsf[t.From.ID]
		if !ok {
			w, err = h.walletFinder.FindByBankAccount(t.From)
			if err != nil {
				if err != mm.ErrNotFound {
					logger.FromContext(ctx).Errorf(
						"An error happen while finding wallet by bank account from [%s] -> error [%s]",
						t.From.AccountNo,
						err,
					)

					return nil, err
				}

				w, ok = wst[t.To.ID]
				if !ok {
					w, err = h.walletFinder.FindByBankAccount(t.To)
					if err != nil {
						logger.FromContext(ctx).Errorf(
							"An error happen while finding wallet by bank account to [%s] -> error [%s]",
							t.To.AccountNo,
							err,
						)

						return nil, err
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

	err = h.transferPersister.PersistAll(ts)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting transfer -> error [%s]",
			err,
		)

		return nil, err
	}

	err = h.walletPersister.UpdateAllAccounting(ws)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting wallets -> error [%s]",
			err,
		)

		return nil, err
	}

	return ts, nil
}
