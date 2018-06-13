package account

import (
	"github.com/satori/go.uuid"

	"time"

	"github.com/dohernandez/market-manager/pkg/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/banking/transfer"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	Service struct {
		walletFinder        wallet.Finder
		walletPersister     wallet.Persister
		stockFinder         stock.Finder
		stockDividendFinder dividend.Finder

		ccClient *cc.Client
	}
)

func NewService(
	walletFinder wallet.Finder,
	walletPersister wallet.Persister,
	stockFinder stock.Finder,
	ccClient *cc.Client,
	stockDividendFinder dividend.Finder,
) *Service {
	return &Service{
		walletFinder:        walletFinder,
		walletPersister:     walletPersister,
		stockFinder:         stockFinder,
		stockDividendFinder: stockDividendFinder,
		ccClient:            ccClient,
	}
}

func (s *Service) SaveAllWallets(ws []*wallet.Wallet) error {
	return s.walletPersister.PersistAll(ws)
}

func (s *Service) SaveAllOperations(w *wallet.Wallet) error {
	cEURUSD, err := s.ccClient.Converter.Get()
	if err != nil {
		return err
	}

	for _, wItem := range w.Items {
		wItem.CapitalRate = cEURUSD.EURUSD

		capital := wItem.Capital()
		w.Capital = capital
	}

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

func (s *Service) UpdateWalletsCapitalByStocks(stks []*stock.Stock) error {
	cEURUSD, err := s.ccClient.Converter.Get()
	if err != nil {
		return err
	}

	for _, stk := range stks {
		ws, err := s.walletFinder.FindWithItemByStock(stk)
		if err != nil {
			if err != mm.ErrNotFound {
				return err
			}

			continue
		}

		for _, w := range ws {
			w.Items[stk.ID].CapitalRate = cEURUSD.EURUSD

			capital := w.Items[stk.ID].Capital()
			w.Capital = capital
		}

		err = s.walletPersister.UpdateAllItemsCapital(ws)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) UpdateWalletsCapitalByStock(stk *stock.Stock) error {
	cEURUSD, err := s.ccClient.Converter.Get()
	if err != nil {
		return err
	}

	ws, err := s.walletFinder.FindWithItemByStock(stk)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}

		return nil
	}

	for _, w := range ws {
		w.Items[stk.ID].CapitalRate = cEURUSD.EURUSD

		capital := w.Items[stk.ID].Capital()
		w.Capital = capital
	}

	err = s.walletPersister.UpdateAllItemsCapital(ws)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) FindWalletWithAllActiveItems(wName string) (*wallet.Wallet, error) {
	w, err := s.walletFinder.FindByName(wName)
	if err != nil {
		return nil, err
	}

	err = s.walletFinder.LoadActiveItems(w)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := s.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return nil, err
		}

		i.Stock = stk

		err = s.walletFinder.LoadItemOperations(i)
		if err != nil {
			return nil, err
		}

		d, err := s.stockDividendFinder.FindDividendNextAnnounceProjectFromYearAndMonth(i.Stock.ID, year, month)
		if err != nil {
			if err != mm.ErrNotFound {
				return nil, err
			}

			continue
		}

		stk.Dividends = append(stk.Dividends, d)
	}

	return w, nil
}
