package handler

import (
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type updateStockPrice struct {
	stockFinder       stock.Finder
	stockPriceService service.StockPrice
	stockPersister    stock.Persister
}

func (h *updateStockPrice) updateStock(stk *stock.Stock) error {
	p, err := h.stockPriceService.Price(stk)
	if err != nil {
		return errors.Wrapf(err, "symbol : %s", stk.Symbol)
	}

	stk.Value = mm.Value{
		Amount:   p.Close,
		Currency: mm.Dollar,
	}

	stk.Change = mm.Value{
		Amount:   p.Change,
		Currency: mm.Dollar,
	}

	if stk.High52week.Amount < p.High {
		stk.High52week = mm.Value{
			Amount:   p.High,
			Currency: stk.High52week.Currency,
		}
	}

	if stk.Low52week.Amount > p.Low {
		stk.Low52week = mm.Value{
			Amount:   p.High,
			Currency: stk.Low52week.Currency,
		}
	}

	if err := h.stockPersister.UpdatePrice(stk); err != nil {
		return errors.Wrapf(err, "symbol : %s", stk.Symbol)
	}

	return nil
}
