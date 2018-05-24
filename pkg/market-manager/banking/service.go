package banking

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type (
	Service struct {
		bankAccountFinder bank.Finder
		transferPersister transfer.Persister
	}
)

func NewService(bankAccountFinder bank.Finder, transferPersister transfer.Persister) *Service {
	return &Service{
		bankAccountFinder: bankAccountFinder,
		transferPersister: transferPersister,
	}
}

func (s *Service) FindBankAccountByAlias(alias string) (*bank.Account, error) {
	return s.bankAccountFinder.FindByAlias(alias)
}

func (s *Service) SaveAllTransfers(ts []*transfer.Transfer) error {
	return s.transferPersister.PersistAll(ts)
}
