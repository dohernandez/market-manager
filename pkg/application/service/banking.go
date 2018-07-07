package service

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type (
	Banking struct {
		bankAccountFinder bank.Finder
		transferPersister transfer.Persister

		accountService *Account
	}
)

func NewBankingService(bankAccountFinder bank.Finder, transferPersister transfer.Persister, accountService *Account) *Banking {
	return &Banking{
		bankAccountFinder: bankAccountFinder,
		transferPersister: transferPersister,
		accountService:    accountService,
	}
}

func (s *Banking) FindBankAccountByAlias(alias string) (*bank.Account, error) {
	return s.bankAccountFinder.FindByAlias(alias)
}

func (s *Banking) SaveAllTransfers(ts []*transfer.Transfer) error {
	err := s.transferPersister.PersistAll(ts)
	if err != nil {
		return err
	}

	// TODO remove all persisted transfer in case error
	return s.accountService.UpdateWalletsAccountingByTransfers(ts)
}
