package banking

import (
	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type (
	Service struct {
		bankAccountFinder bank.Finder
		transferPersister transfer.Persister

		accountService *account.Service
	}
)

func NewService(bankAccountFinder bank.Finder, transferPersister transfer.Persister, accountService *account.Service) *Service {
	return &Service{
		bankAccountFinder: bankAccountFinder,
		transferPersister: transferPersister,
		accountService:    accountService,
	}
}

func (s *Service) FindBankAccountByAlias(alias string) (*bank.Account, error) {
	return s.bankAccountFinder.FindByAlias(alias)
}

func (s *Service) SaveAllTransfers(ts []*transfer.Transfer) error {
	err := s.transferPersister.PersistAll(ts)
	if err != nil {
		return err
	}

	// TODO remove all persisted transfer in case error
	return s.updateWalletAccounting(ts)
}

func (s *Service) updateWalletAccounting(ts []*transfer.Transfer) error {
	var ws []*wallet.Wallet

	wsf := map[uuid.UUID]*wallet.Wallet{}
	wst := map[uuid.UUID]*wallet.Wallet{}

	for _, t := range ts {
		var err error

		w, ok := wsf[t.From.ID]
		if !ok {
			w, err = s.accountService.FindWalletByBankAccount(t.From)
			if err != nil {
				if err != mm.ErrNotFound {
					return err
				}

				w, ok = wst[t.To.ID]
				if !ok {
					w, err = s.accountService.FindWalletByBankAccount(t.To)
					if err != nil {
						if err != mm.ErrNotFound {
							return err
						}

						continue
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

	return s.accountService.UpdateAllWalletsAccounting(ws)
}
