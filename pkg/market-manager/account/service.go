package account

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/bank"
)

type (
	Service struct {
		walletFinder    wallet.Finder
		walletPersister wallet.Persister
	}
)

func NewService(walletFinder wallet.Finder, walletPersister wallet.Persister) *Service {
	return &Service{
		walletFinder:    walletFinder,
		walletPersister: walletPersister,
	}
}

func (s *Service) SaveAllWallets(ws []*wallet.Wallet) error {
	return s.walletPersister.PersistAll(ws)
}

func (s *Service) SaveAllOperations(w *wallet.Wallet) error {
	return s.walletPersister.PersistOperations(w)
}

func (s *Service) FindWalletByName(name string) (*wallet.Wallet, error) {
	return s.walletFinder.FindByName(name)
}

func (s *Service) FindWalletByBankAccount(ba *bank.Account) (*wallet.Wallet, error) {
	return s.walletFinder.FindByBankAccount(ba)
}

func (s *Service) UpdateAllWalletsAccounting(ws []*wallet.Wallet) error {
	return s.walletPersister.UpdateAllAccounting(ws)
}

//func (s *Service) FindWalletItem(stk *stock.Stock) (*wallet.Item, error) {
//	//return s.walletFinder.FindByStock(stk)
//	return nil, mm.ErrNotFound
//}
//
//func (s *Service) SaveAllWalletItem(is []*wallet.Item) error {
//	for _, i := range is {
//		fmt.Printf("%+v\n", i)
//	}
//	//return s.itemPersister.PersistAll(is)
//	return nil
//}
//
//func (s *Service) BuyStock(o *operation.Operation) error {
//	i, err := s.walletFinder.FindByStock(o.Stock)
//	if err != nil {
//		if err != mm.ErrNotFound {
//			return err
//		}
//
//		i = wallet.NewItem(o.Stock)
//	}
//
//	i.increaseInvestment(o.Amount, o.Price)
//	err = s.itemPersister.Persist(i)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (s *Service) SellStock(o *operation.Operation) error {
//	i, err := s.walletFinder.FindByStock(o.Stock)
//	if err != nil {
//		return err
//	}
//
//	i.decreaseInvestment(o.Amount, o.Price)
//	err = s.itemPersister.Persist(i)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (s *Service) Dividend(o *operation.Operation) error {
//	i, err := s.walletFinder.FindByStock(o.Stock)
//	if err != nil {
//		return err
//	}
//
//	i.increaseDividend(o.Price)
//	err = s.itemPersister.Persist(i)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
