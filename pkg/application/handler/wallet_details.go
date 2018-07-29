package handler

import (
	"context"
	"time"

	"github.com/gogolfing/cbus"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/render"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/currency-converter"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/account/wallet"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type walletDetails struct {
	walletFinder   wallet.Finder
	stockFinder    stock.Finder
	dividendFinder dividend.Finder
	ccClient       *cc.Client
	retention      float64
}

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
	wName := command.(*appCommand.WalletDetails).Wallet

	w, err := h.LoadWalletWithActiveWalletItems(wName)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while loading wallet [%s] -> error [%s]",
			wName,
			err,
		)

		return nil, err
	}

	wDProjectedGrossMonth := w.DividendProjectedNextMonth()

	wDProjectedMonth := wDProjectedGrossMonth.Decrease(mm.Value{
		Amount:   h.retention * wDProjectedGrossMonth.Amount / 100,
		Currency: wDProjectedGrossMonth.Currency,
	})

	wDProjectedGrossYear := w.DividendProjectedNextYear()

	wDProjectedYear := wDProjectedGrossYear.Decrease(mm.Value{
		Amount:   h.retention * wDProjectedGrossYear.Amount / 100,
		Currency: wDProjectedGrossYear.Currency,
	})

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
			DividendYearProjected:  wDProjectedYear,

			Connection: w.Connection,
			Interest:   w.Interest,
			Commission: w.Commission,
		},
	}

	//WalletOutput struct {
	//
	//	DYield             float64
	//}

	var wSOutputs []*render.WalletStockOutput
	for _, item := range w.Items {
		var (
			exDate   time.Time
			wADYield float64
			dividend mm.Value
		)

		wAPrice := item.WeightedAveragePrice()

		if len(item.Stock.Dividends) > 0 {
			d := item.Stock.Dividends[0]
			exDate = d.ExDate

			if d.Amount.Amount > 0 {
				dividend = d.Amount
				wADYield = d.Amount.Amount * 4 / wAPrice.Amount * 100
			}
		}

		wSOutputs = append(wSOutputs, &render.WalletStockOutput{
			StockOutput: render.StockOutput{
				Stock:      item.Stock.Name,
				Market:     item.Stock.Exchange.Symbol,
				Symbol:     item.Stock.Symbol,
				Value:      item.Stock.Value,
				High52Week: item.Stock.High52Week,
				Low52Week:  item.Stock.Low52Week,
				BuyUnder:   item.Stock.BuyUnder(),
				ExDate:     exDate,
				DYield:     item.Stock.DividendYield,
				EPS:        item.Stock.EPS,
				Change:     item.Stock.Change,
				UpdatedAt:  item.Stock.LastPriceUpdate,

				PriceWithHighLow: item.Stock.ComparePriceWithHighLow(),
			},
			Amount:             item.Amount,
			Capital:            item.Capital(),
			Invested:           item.Invested,
			Dividend:           dividend,
			DividendPayed:      item.Dividend,
			PercentageWallet:   item.PercentageInvestedRepresented(w.Capital.Amount),
			Buys:               item.Buys,
			Sells:              item.Sells,
			NetBenefits:        item.NetBenefits(),
			PercentageBenefits: item.PercentageBenefits(),
			Change:             item.Change(),
			WAPrice:            wAPrice,
			WADYield:           wADYield,
		})
	}

	wDetailsOutput.WalletStockOutputs = wSOutputs

	return wDetailsOutput, err
}

func (h *walletDetails) LoadWalletWithActiveWalletItems(name string) (*wallet.Wallet, error) {
	w, err := h.walletFinder.FindByName(name)
	if err != nil {
		return nil, err
	}

	err = h.walletFinder.LoadActiveItems(w)
	if err != nil {
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

		i.Stock = stk

		err = h.walletFinder.LoadItemOperations(i)
		if err != nil {
			return nil, err
		}

		ds, err := h.dividendFinder.FindAllDividendsFromThisYearAndMontOn(i.Stock.ID, year, month)
		i.Stock.Dividends = ds
	}

	cEURUSD, err := h.ccClient.Converter.Get()
	if err != nil {
		return nil, err
	}

	w.SetCapitalRate(cEURUSD.EURUSD)

	return w, err
}
