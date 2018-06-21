package account

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
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

func (s *Service) LoadActiveWalletItems(w *wallet.Wallet) error {
	err := s.walletFinder.LoadActiveItems(w)
	if err != nil {
		return err
	}

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := s.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return err
		}

		i.Stock = stk

		err = s.walletFinder.LoadItemOperations(i)
		if err != nil {
			return err
		}

		ds, err := s.stockDividendFinder.FindAllFormStock(i.Stock.ID)
		if err != nil {
			if err != mm.ErrNotFound {
				return err
			}

			continue
		}

		for _, d := range ds {
			stk.Dividends = append(stk.Dividends, d)
		}
	}

	return nil
}

func (s *Service) LoadWalletItem(w *wallet.Wallet, stkSymbol string) error {
	stk, err := s.stockFinder.FindBySymbol(stkSymbol)
	if err != nil {
		return err
	}

	err = s.walletFinder.LoadItemByStock(w, stk)
	if err != nil {
		return err
	}

	i, _ := w.Items[stk.ID]
	err = s.walletFinder.LoadItemOperations(i)
	if err != nil {
		return err
	}

	ds, err := s.stockDividendFinder.FindAllFormStock(i.Stock.ID)
	if err != nil {
		if err != mm.ErrNotFound {
			return err
		}
	}

	for _, d := range ds {
		stk.Dividends = append(stk.Dividends, d)
	}

	return nil
}

type AppCommissions struct {
	Commission struct {
		Base struct {
			Amount   float64
			Currency string
		}
		Extra struct {
			Amount   float64
			Currency string
			Apply    string
		}
		Maximum struct {
			Amount   float64
			Currency string
		}
	}
	ChangeCommission struct {
		Amount   float64
		Currency string
	}
}

func (s *Service) SellStocksWallet(
	w *wallet.Wallet,
	stksSymbol map[string]int,
	pChangeCommissions map[string]mm.Value,
	commissions map[string]AppCommissions,
) error {
	for symbol, amount := range stksSymbol {
		stk, err := s.stockFinder.FindBySymbol(symbol)
		if err != nil {
			return err
		}

		o := s.createOperation(stk, amount, operation.Sell, w.CurrentCapitalRate(), pChangeCommissions, commissions)
		w.AddOperation(o)
	}

	return nil
}

func (s *Service) createOperation(
	stk *stock.Stock,
	amount int,
	action operation.Action,
	capitalRate float64,
	pChangeCommissions map[string]mm.Value,
	commissions map[string]AppCommissions,
) *operation.Operation {
	pChange := mm.Value{
		Amount:   capitalRate,
		Currency: mm.Dollar,
	}

	now := time.Now()

	oValue := mm.Value{
		Amount:   stk.Value.Amount * float64(amount) / pChange.Amount,
		Currency: mm.Euro,
	}

	pChangeCommission, _ := pChangeCommissions[stk.Exchange.Symbol]

	var commission mm.Value
	appCommission, ok := commissions[stk.Exchange.Symbol]
	if !ok {
		if stk.Exchange.Symbol == "NASDAQ" || stk.Exchange.Symbol == "NYSE" {
			commission.Amount = appCommission.Commission.Base.Amount
			extra := appCommission.Commission.Extra.Amount * float64(amount) / pChange.Amount

			commission = commission.Increase(mm.Value{Amount: extra, Currency: mm.Euro})
		} else {
			panic("Commission to apply not defined")
		}
	}

	o := operation.NewOperation(now, stk, action, amount, stk.Value, pChange, pChangeCommission, oValue, commission)

	return o
}

func (s *Service) BuyStocksWallet(
	w *wallet.Wallet,
	stksSymbol map[string]int,
	pChangeCommissions map[string]mm.Value,
	commissions map[string]AppCommissions,
) error {
	for symbol, amount := range stksSymbol {
		stk, err := s.stockFinder.FindBySymbol(symbol)
		if err != nil {
			return err
		}

		o := s.createOperation(stk, amount, operation.Buy, w.CurrentCapitalRate(), pChangeCommissions, commissions)

		ds, err := s.stockDividendFinder.FindAllFormStock(o.Stock.ID)
		if err != nil {
			if err != mm.ErrNotFound {
				return err
			}
		}

		for _, d := range ds {
			o.Stock.Dividends = append(o.Stock.Dividends, d)
		}

		w.AddOperation(o)
	}

	return nil
}
