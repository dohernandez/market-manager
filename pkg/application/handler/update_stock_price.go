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

	if p.High52Week > 0 {
		stk.High52Week = mm.Value{
			Amount:   p.High52Week,
			Currency: stk.High52Week.Currency,
		}
	} else if stk.High52Week.Amount < p.High {
		stk.High52Week = mm.Value{
			Amount:   p.High,
			Currency: stk.High52Week.Currency,
		}
	}

	if p.Low52Week > 0 {
		stk.Low52Week = mm.Value{
			Amount:   p.Low52Week,
			Currency: stk.Low52Week.Currency,
		}
	} else if stk.Low52Week.Amount > p.Low {
		stk.Low52Week = mm.Value{
			Amount:   p.High,
			Currency: stk.Low52Week.Currency,
		}
	}

	stk.EPS = p.EPS
	stk.PER = p.PER

	if err := h.stockPersister.UpdatePrice(stk); err != nil {
		return errors.Wrapf(err, "symbol : %s", stk.Symbol)
	}

	return nil
}
