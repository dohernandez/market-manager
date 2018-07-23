package handler

import (
	"context"

	"github.com/gogolfing/cbus"

	"io"

	appCommand "github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type importStock struct {
	marketFinder       market.Finder
	exchangeFinder     exchange.Finder
	stockInfoFinder    stock.InfoFinder
	stockInfoPersister stock.InfoPersister

	stockInfos map[string]*stock.Info
}

func NewImportStock(
	marketFinder market.Finder,
	exchangeFinder exchange.Finder,
	stockInfoFinder stock.InfoFinder,
	stockInfoPersister stock.InfoPersister,
) *importStock {
	return &importStock{
		marketFinder:       marketFinder,
		exchangeFinder:     exchangeFinder,
		stockInfoFinder:    stockInfoFinder,
		stockInfoPersister: stockInfoPersister,
		stockInfos:         map[string]*stock.Info{},
	}
}

func (h *importStock) Handle(ctx context.Context, command cbus.Command) (result interface{}, err error) {
	filePath := command.(*appCommand.ImportStock).FilePath
	r := util.NewCsvReader(filePath)

	r.Open()
	defer r.Close()

	var ss []*stock.Stock

	for {
		line, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(ctx).Fatal(err)
		}

		m, err := h.marketFinder.FindByName(market.Stock)
		if err != nil {
			return nil, err
		}

		e, err := h.exchangeFinder.FindBySymbol(line[1])
		if err != nil {
			return nil, err
		}

		t, err := h.getStockInfo(ctx, line[3], stock.StockInfoType)
		if err != nil {
			return nil, err
		}

		sector, err := h.getStockInfo(ctx, line[4], stock.StockInfoSector)
		if err != nil {
			return nil, err
		}

		industry, err := h.getStockInfo(ctx, line[5], stock.StockInfoIndustry)
		if err != nil {
			return nil, err
		}

		stk := stock.NewStock(m, e, line[0], line[2], t, sector, industry)
		ss = append(ss, stk)

		logger.FromContext(ctx).Debugf("Added new stock [%+v]", stk)
	}

	return ss, nil
}

func (h *importStock) getStockInfo(ctx context.Context, value string, t stock.InfoType) (*stock.Info, error) {
	if stkInfo, ok := h.stockInfos[value]; ok {
		return stkInfo, nil
	}

	stkInfo, err := h.stockInfoFinder.FindByName(value)
	if err != nil {
		if err != mm.ErrNotFound {
			return nil, err
		}

		stkInfo = stock.NewStockInfo(value, t)

		err = h.stockInfoPersister.Persist(stkInfo)
		if err != nil {
			return nil, err
		}

		logger.FromContext(ctx).Debugf("Persisted a new stock info [%+v]", stkInfo)

		h.stockInfos[value] = stkInfo
	}

	return stkInfo, nil
}
