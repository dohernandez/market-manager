package account

import (
	"fmt"

	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
)

type (
	Service struct {
		operationPersister operation.Persister
		itemFinder         wallet.Finder
		itemPersister      wallet.Persister
	}
)

func NewService(accountPersister operation.Persister, itemFinder wallet.Finder, itemPersister wallet.Persister) *Service {
	return &Service{
		operationPersister: accountPersister,
		itemFinder:         itemFinder,
		itemPersister:      itemPersister,
	}
}

func (s *Service) SaveAllOperations(os []*operation.Operation) error {
	for _, o := range os {
		fmt.Printf("%+v\n", o)
	}
	//return s.operationPersister.PersistAll(os)

	//for _, o := range os {
	//
	//}

	//if action == operation.Buy || action == operation.Sell || action == operation.Dividend {
	//	wi, ok := cis[s.ID]
	//	if !ok {
	//		wi, err = i.walletService.FindWalletItem(s)
	//		if err != nil {
	//			if err != mm.ErrNotFound {
	//				return errors.New(fmt.Sprintf("find wallet item for stock %s: %s", line[2], err.Error()))
	//			}
	//
	//			wi = new(wallet.Item)
	//			wi.Stock = s
	//
	//			cis[s.ID] = wi
	//		}
	//	}
	//
	//	switch action {
	//	case operation.Buy:
	//		wi.IncreaseInvestment(amount, value)
	//	case operation.Sell:
	//		wi.DecreaseInvestment(amount, value)
	//	case operation.Dividend:
	//		wi.IncreaseDividend(value)
	//	}
	//}
	return nil
}

//func (s *Service) FindWalletItem(stk *stock.Stock) (*wallet.Item, error) {
//	//return s.itemFinder.FindByStock(stk)
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
//	i, err := s.itemFinder.FindByStock(o.Stock)
//	if err != nil {
//		if err != mm.ErrNotFound {
//			return err
//		}
//
//		i = wallet.NewItem(o.Stock)
//	}
//
//	i.IncreaseInvestment(o.Amount, o.Price)
//	err = s.itemPersister.Persist(i)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (s *Service) SellStock(o *operation.Operation) error {
//	i, err := s.itemFinder.FindByStock(o.Stock)
//	if err != nil {
//		return err
//	}
//
//	i.DecreaseInvestment(o.Amount, o.Price)
//	err = s.itemPersister.Persist(i)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (s *Service) Dividend(o *operation.Operation) error {
//	i, err := s.itemFinder.FindByStock(o.Stock)
//	if err != nil {
//		return err
//	}
//
//	i.IncreaseDividend(o.Price)
//	err = s.itemPersister.Persist(i)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
