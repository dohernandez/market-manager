package listener

import (
	"context"

	"github.com/gogolfing/cbus"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type addStockSummaryInfo struct {
	stockSummaryMarketChameleon service.StockSummary
	stockSummaryYahoo           service.StockSummary
}

func NewAddStockSummaryInfo(stockSummaryMarketChameleon service.StockSummary, stockSummaryYahoo service.StockSummary) *addStockSummaryInfo {
	return &addStockSummaryInfo{
		stockSummaryMarketChameleon: stockSummaryMarketChameleon,
		stockSummaryYahoo:           stockSummaryYahoo,
	}
}

func (l *addStockSummaryInfo) OnEvent(ctx context.Context, event cbus.Event) {
	s, ok := event.Result.(*stock.Stock)
	if !ok {
		logger.FromContext(ctx).Warn("addStockSummaryInfo: Result instance not supported")

		return
	}

	summaryMarketChameleon, err := l.stockSummaryMarketChameleon.Summary(s)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"addStockSummaryInfo: can load summary MarketChameleon for stock [%s]",
			s.Symbol,
		)

	}

	summaryYahoo, err := l.stockSummaryYahoo.Summary(s)
	if err != nil {
		logger.FromContext(ctx).Errorf(
			"addStockSummaryInfo: can load summary Yahoo for stock [%s]",
			s.Symbol,
		)

	}

	s.Name = summaryYahoo.Name

	s.Type = stock.NewStockInfo(summaryMarketChameleon.Type, stock.StockInfoType)
	s.Industry = stock.NewStockInfo(summaryMarketChameleon.Type, stock.StockInfoIndustry)
	s.Sector = stock.NewStockInfo(summaryMarketChameleon.Type, stock.StockInfoSector)

	return
}
