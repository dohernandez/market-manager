package banking

import (
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
)

type (
	Service struct {
	}
)

func NewService() *Service {
	return &Service{}
}

func (s *Service) FindBankAccountByAlias(alias string) (*bank.Account, error) {
	return nil, mm.ErrNotFound
}

func (s *Service) SaveAllTransfers(ts []*transfer.Transfer) error {
	return nil
}
