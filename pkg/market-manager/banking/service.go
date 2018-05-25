package banking

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/account"
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
	return s.accountService.UpdateWalletsAccountingByTransfers(ts)
}
