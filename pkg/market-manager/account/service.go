package account

import (
	"errors"
	"fmt"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
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
	err := s.operationPersister.PersistAll(os)
	if err != nil {
		return err
	}

	cwis := make(map[uuid.UUID]*wallet.Item)
	for _, o := range os {
		if o.Action == operation.Buy || o.Action == operation.Sell || o.Action == operation.Dividend {
			wi, ok := cwis[o.Stock.ID]
			if !ok {
				wi, err = s.itemFinder.FindByStock(o.Stock)
				if err != nil {
					if err != mm.ErrNotFound {
						return errors.New(fmt.Sprintf("find wallet item for stock %s: %s", o.Stock.Symbol, err.Error()))
					}

					wi = wallet.NewItem(o.Stock)
				}
				cwis[o.Stock.ID] = wi
			}

			switch o.Action {
			case operation.Buy:
				wi.IncreaseInvestment(o.Amount, o.Value, o.PriceChangeCommission, o.Commission)
			case operation.Sell:
				wi.DecreaseInvestment(o.Amount, o.Value, o.PriceChangeCommission, o.Commission)
			case operation.Dividend:
				wi.IncreaseDividend(o.Value)
			}
		}
	}

	var wis []*wallet.Item
	for _, wi := range cwis {
		wis = append(wis, wi)
	}

	return s.itemPersister.PersistAll(wis)
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
