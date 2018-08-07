package handler

import (
	"context"
	"time"

	"github.com/gogolfing/cbus"

	"errors"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/render"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/operation"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	walletDetails struct {
		walletFinder   wallet.Finder
		stockFinder    stock.Finder
		dividendFinder dividend.Finder
		ccClient       *cc.Client
		retention      float64
	}
)

func NewWalletDetails(
	walletFinder wallet.Finder,
	stockFinder stock.Finder,
	dividendFinder dividend.Finder,
	ccClient *cc.Client,
	retention float64,
) *walletDetails {
	return &walletDetails{
		walletFinder:   walletFinder,
		stockFinder:    stockFinder,
		dividendFinder: dividendFinder,
		ccClient:       ccClient,
		retention:      retention,
	}
}

func (h *walletDetails) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	walletDetails := command.(*appCommand.WalletDetails)

	wName := walletDetails.Wallet
	if wName == "" {
		logger.FromContext(ctx).Error("An error happen while loading wallet -> error [wallet can not be empty]")

		return nil, errors.New("missing wallet name")
	}

	sells := walletDetails.Sells
	buys := walletDetails.Buys
	commissions := walletDetails.Commissions
	status := walletDetails.Status

	w, err := h.loadWalletWithWalletItemsAndWalletTrades(wName, status)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	if len(sells) > 0 {
		for symbol, amount := range sells {
			stk, err := h.stockFinder.FindBySymbol(symbol)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while loading sell stock symbol [%s] -> error [%s]",
					wName,
					err,
				)

				return nil, err
			}

			o := h.createOperation(stk, amount, operation.Sell, w.CurrentCapitalRate(), commissions)

			w.AddOperation(o)
		}
	}

	if len(buys) > 0 {
		now := time.Now()
		month := int(now.Month())
		year := now.Year()

		for symbol, amount := range buys {
			stk, err := h.stockFinder.FindBySymbol(symbol)
			if err != nil {
				logger.FromContext(ctx).Errorf(
					"An error happen while loading buys stock symbol [%s] -> error [%s]",
					wName,
					err,
				)

				return nil, err
			}

			o := h.createOperation(stk, amount, operation.Buy, w.CurrentCapitalRate(), commissions)

			ds, err := h.dividendFinder.FindAllDividendsFromThisYearAndMontOn(stk.ID, year, month)
			if err != nil {
				if err != mm.ErrNotFound {
					logger.FromContext(ctx).Errorf(
						"An error happen while loading dividends for stock bought symbol [%s] -> error [%s]",
						stk.Symbol,
						err,
					)

					return nil, err
				}
			}

			o.Stock.Dividends = ds

			w.AddOperation(o)
		}

		if w.FreeMargin().Amount < 0 {
			logger.FromContext(ctx).Errorf(
				"An error happen there is not enough funds to execute the buys wallet [%s]",
				wName,
			)

			return nil, errors.New("not enough funds to execute the buys")
		}
	}

	wDProjectedGrossMonth := w.DividendProjectedNextMonth()

	wDProjectedMonth := wDProjectedGrossMonth.Decrease(mm.Value{
		Amount:   h.retention * wDProjectedGrossMonth.Amount / 100,
		Currency: wDProjectedGrossMonth.Currency,
	})

	dividendMonthYield := wDProjectedMonth.Amount * 100 / w.Invested.Amount

	wDProjectedGrossYear := w.DividendProjectedNextYear()

	wDProjectedYear := wDProjectedGrossYear.Decrease(mm.Value{
		Amount:   h.retention * wDProjectedGrossYear.Amount / 100,
		Currency: wDProjectedGrossYear.Currency,
	})

	dividendYearYield := wDProjectedYear.Amount * 100 / w.Invested.Amount

	wDetailsOutput := render.WalletDetailsOutput{
		WalletOutput: render.WalletOutput{
			Capital:                w.Capital,
			Invested:               w.Invested,
			Funds:                  w.Funds,
			FreeMargin:             w.FreeMargin(),
			NetCapital:             w.NetCapital(),
			NetBenefits:            w.NetBenefits(),
			PercentageBenefits:     w.PercentageBenefits(),
			DividendPayed:          w.Dividend,
			DividendMonthProjected: wDProjectedMonth,
			DividendMonthYield:     dividendMonthYield,
			DividendYearProjected:  wDProjectedYear,
			DividendYearYield:      dividendYearYield,
			Connection:             w.Connection,
			Interest:               w.Interest,
			Commission:             w.Commission,
		},
	}

	var wSOutputs []*render.WalletStockOutput
	for _, item := range w.Items {
		if status == operation.Active && item.Amount == 0 {
			continue
		}

		if status == operation.Inactive && item.Amount != 0 {
			continue
		}

		var (
			exDate          time.Time
			wADYield        float64
			sDividend       mm.Value
			sDividendStatus dividend.Status
			dividendToPay   mm.Value
		)

		wAPrice := item.WeightedAveragePrice()

		if len(item.Stock.Dividends) > 0 {
			d := item.Stock.Dividends[0]
			exDate = d.ExDate

			if d.Amount.Amount > 0 {
				sDividend = d.Amount
				wADYield = d.Amount.Amount * 4 / wAPrice.Amount * 100

				sDividendStatus = d.Status

				dividendToPayGross := mm.Value{
					Amount:   float64(item.Amount) * d.Amount.Amount,
					Currency: mm.Dollar,
				}

				dividendToPay = mm.Value{
					Amount:   (dividendToPayGross.Amount - (h.retention * dividendToPayGross.Amount / 100)) / w.CurrentCapitalRate(),
					Currency: mm.Euro,
				}
			}
		}

		var sTrades []*render.TradeOutput

		for _, t := range item.Trades {
			var isProfitable bool

			if t.Net().Amount > 0 {
				isProfitable = true
			}

			sTrades = append(sTrades, &render.TradeOutput{
				Stock:  t.Stock.Name,
				Market: t.Stock.Exchange.Symbol,
				Symbol: t.Stock.Symbol,
				Enter: struct {
					Amount float64
					Kurs   mm.Value
					Total  mm.Value
				}{Amount: t.BuysAmount, Kurs: mm.Value{}, Total: t.Buys},
				Position: struct {
					Amount   float64
					Dividend mm.Value
					Capital  mm.Value
				}{Amount: t.Amount, Dividend: t.Dividend, Capital: t.Capital()},
				Exit: struct {
					Amount float64
					Kurs   mm.Value
					Total  mm.Value
				}{Amount: t.SellsAmount, Kurs: mm.Value{}, Total: t.Sells},

				BenefitPercentage: t.BenefitPercentage(),
				Net:               t.Net(),
				IsProfitable:      isProfitable,
			})
		}

		wSOutputs = append(wSOutputs, &render.WalletStockOutput{
			StockOutput: render.StockOutput{
				Stock:          item.Stock.Name,
				Market:         item.Stock.Exchange.Symbol,
				Symbol:         item.Stock.Symbol,
				Value:          item.Stock.Value,
				High52Week:     item.Stock.High52Week,
				Low52Week:      item.Stock.Low52Week,
				BuyUnder:       item.Stock.BuyUnder(),
				ExDate:         exDate,
				Dividend:       sDividend,
				DYield:         item.Stock.DividendYield,
				DividendStatus: sDividendStatus,
				EPS:            item.Stock.EPS,
				Change:         item.Stock.Change,
				UpdatedAt:      item.Stock.LastPriceUpdate,
				HV52Week:       item.Stock.HV52Week,
				HV20Day:        item.Stock.HV20Day,

				PriceWithHighLow: item.Stock.ComparePriceWithHighLow(),
			},
			Amount:             item.Amount,
			Capital:            item.Capital(),
			Invested:           item.Invested,
			DividendPayed:      item.Dividend,
			DividendToPay:      dividendToPay,
			PercentageWallet:   item.PercentageInvestedRepresented(w.Capital.Amount),
			Buys:               item.Buys,
			Sells:              item.Sells,
			NetBenefits:        item.NetBenefits(),
			PercentageBenefits: item.PercentageBenefits(),
			Change:             item.Change(),
			WAPrice:            wAPrice,
			WADYield:           wADYield,
			Trades:             sTrades,
		})
	}

	wDetailsOutput.WalletStockOutputs = wSOutputs

	return wDetailsOutput, err
}

func (h *walletDetails) loadWalletWithWalletItemsAndWalletTrades(name string, status operation.Status) (*wallet.Wallet, error) {
	w, err := h.walletFinder.FindByName(name)
	if err != nil {
		return nil, err
	}

	switch status {
	case operation.Inactive:
		if err = h.walletFinder.LoadInactiveItems(w); err != nil {
			return nil, err
		}
	case operation.All:
		if err = h.walletFinder.LoadAllItems(w); err != nil {
			return nil, err
		}
	default:
		if err = h.walletFinder.LoadActiveItems(w); err != nil {
			return nil, err
		}
	}

	if err = h.walletFinder.LoadActiveTrades(w); err != nil {
		return nil, err
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	for _, i := range w.Items {
		// Add this into go routing. Use the example explain in the page
		// https://medium.com/@trevor4e/learning-gos-concurrency-through-illustrations-8c4aff603b3
		stk, err := h.stockFinder.FindByID(i.Stock.ID)
		if err != nil {
			return nil, err
		}

		// I like to keep the address but change the content to keep,
		// trade and item pointing to the same stock
		*i.Stock = *stk

		err = h.walletFinder.LoadItemOperations(i)
		if err != nil {
			return nil, err
		}

		ds, err := h.dividendFinder.FindAllDividendsFromThisYearAndMontOn(i.Stock.ID, year, month)
		if err != nil {
			if err != mm.ErrNotFound {
				return nil, err
			}
		}

		i.Stock.Dividends = ds
	}

	cEURUSD, err := h.ccClient.Converter.Get()
	if err != nil {
		return nil, err
	}

	w.SetCapitalRate(cEURUSD.EURUSD)

	return w, err
}

func (h *walletDetails) createOperation(
	stk *stock.Stock,
	amount int,
	action operation.Action,
	capitalRate float64,
	commissions map[string]appCommand.Commission,
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

	var commission, pChangeCommission mm.Value
	marketCommission, ok := commissions[stk.Exchange.Symbol]
	if ok {
		pChangeCommission.Amount = marketCommission.ChangeCommission.Amount
		pChangeCommission.Currency = mm.Currency(marketCommission.ChangeCommission.Currency)

		if stk.Exchange.Symbol == "NASDAQ" || stk.Exchange.Symbol == "NYSE" {
			commission.Amount = marketCommission.Commission.Base.Amount
			extra := marketCommission.Commission.Extra.Amount * float64(amount) / pChange.Amount

			commission = commission.Increase(mm.Value{Amount: extra, Currency: mm.Euro})
		} else {
			panic("Commission to apply not defined")
		}
	}

	o := operation.NewOperation(now, stk, action, amount, stk.Value, pChange, pChangeCommission, oValue, commission)

	return o
}
