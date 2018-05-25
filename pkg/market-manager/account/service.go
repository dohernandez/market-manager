package account

import (
	uuid "github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
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

func (s *Service) UpdateWalletsAccountingByTransfers(ts []*transfer.Transfer) error {
	var ws []*wallet.Wallet

	wsf := map[uuid.UUID]*wallet.Wallet{}
	wst := map[uuid.UUID]*wallet.Wallet{}

	for _, t := range ts {
		var err error

		w, ok := wsf[t.From.ID]
		if !ok {
			w, err = s.walletFinder.FindByBankAccount(t.From)
			if err != nil {
				if err != mm.ErrNotFound {
					return err
				}

				w, ok = wst[t.To.ID]
				if !ok {
					w, err = s.walletFinder.FindByBankAccount(t.To)
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

	return s.walletPersister.UpdateAllAccounting(ws)
}

func (s *Service) UpdateWalletsAccountingByStocks(stks []*stock.Stock) error {
	var ws []*wallet.Wallet
	cwms := map[uuid.UUID][]*wallet.Wallet{}

	for _, stk := range stks {
		var err error
		rws, ok := cwms[stk.ID]
		if !ok {
			rws, err = s.walletFinder.FindWalletsByStock(stk)
			if err != nil {
				if err != mm.ErrNotFound {
					return err
				}

				continue
			}

			cwms[stk.ID] = rws

			for _, w := range rws {
				ws = append(ws, w)
			}
		}

		for _, w := range ws {
			w.IncreaseCapital(stk.Change)
		}
	}

	return s.walletPersister.UpdateAllAccounting(ws)
}

func (s *Service) UpdateWalletsAccountingByStock(stk *stock.Stock) error {
	ws, err := s.walletFinder.FindWalletsByStock(stk)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		return nil
	}

	for _, w := range ws {
		w.IncreaseCapital(stk.Change)
	}

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
