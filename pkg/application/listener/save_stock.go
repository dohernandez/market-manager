package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type saveStock struct {
	stockInfoFinder    stock.InfoFinder
	stockPersister     stock.Persister
	stockInfoPersister stock.InfoPersister

	stockInfos map[string]*stock.Info
}

func NewSaveStock(
	stockInfoFinder stock.InfoFinder,
	stockPersister stock.Persister,
	stockInfoPersister stock.InfoPersister,
) *saveStock {
	return &saveStock{
		stockInfoFinder:    stockInfoFinder,
		stockPersister:     stockPersister,
		stockInfoPersister: stockInfoPersister,
		stockInfos:         map[string]*stock.Info{},
	}
}

func (l *saveStock) OnEvent(ctx context.Context, event cbus.Event) {
	ss, ok := event.Result.([]*stock.Stock)
	if !ok {
		logger.FromContext(ctx).Warn("updateWalletCapital: Result instance not supported")

		return
	}

	for _, s := range ss {
		err := l.saveAllStockInfo(ctx, s)
		if err != nil {
			logger.FromContext(ctx).Errorf(
				"An error happen while persisting stocks info -> error [%s]",
				err,
			)

			return
		}
	}

	err := l.stockPersister.PersistAll(ss)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"An error happen while persisting stocks -> error [%s]",
			err,
		)

		return
	}

	return
}

func (l *saveStock) saveAllStockInfo(ctx context.Context, s *stock.Stock) error {
	typeStockInfo, err := l.createOrLoadStockInfo(ctx, s.Type.Name, stock.StockInfoType)
	if err != nil {
		return err
	}

	s.Type = typeStockInfo

	sectorStockInfo, err := l.createOrLoadStockInfo(ctx, s.Sector.Name, stock.StockInfoSector)
	if err != nil {
		return err
	}

	s.Sector = sectorStockInfo

	industryStockInfo, err := l.createOrLoadStockInfo(ctx, s.Industry.Name, stock.StockInfoIndustry)
	if err != nil {
		return err
	}

	s.Industry = industryStockInfo

	return nil
}

func (l *saveStock) createOrLoadStockInfo(ctx context.Context, value string, t stock.InfoType) (*stock.Info, error) {
	if stkInfo, ok := l.stockInfos[value]; ok {
		return stkInfo, nil
	}

	stkInfo, err := l.stockInfoFinder.FindByName(value)
	if err != nil {
		if err != mm.ErrNotFound {
			return nil, err
		}

		stkInfo = stock.NewStockInfo(value, t)

		err = l.stockInfoPersister.Persist(stkInfo)
		if err != nil {
			return nil, err
		}

		logger.FromContext(ctx).Debugf("Persisted a new stock info [%+v]", stkInfo)

		l.stockInfos[value] = stkInfo
	}

	return stkInfo, nil
}
