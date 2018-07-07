package import_purchase

import (
	"context"
	"io"

	"github.com/dohernandez/market-manager/pkg/application/service"
	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	ImportStock struct {
		ctx             context.Context
		reader          _import.Reader
		purchaseService *service.Purchase

		stockInfos map[string]*stock.Info
	}
)

func NewImportStock(
	ctx context.Context,
	reader _import.Reader,
	purchaseService *service.Purchase,
) *ImportStock {
	return &ImportStock{
		ctx:             ctx,
		reader:          reader,
		purchaseService: purchaseService,
		stockInfos:      map[string]*stock.Info{},
	}
}

func (i *ImportStock) Import() error {
	i.reader.Open()
	defer i.reader.Close()

	var ss []*stock.Stock

	for {
		line, err := i.reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			logger.FromContext(i.ctx).Fatal(err)
		}

		m, err := i.purchaseService.FindMarketByName(market.Stock)
		if err != nil {
			return err
		}

		e, err := i.purchaseService.FindExchangeBySymbol(line[1])
		if err != nil {
			return err
		}

		t, err := i.getStockInfo(line[3], stock.StockInfoType)
		if err != nil {
			return err
		}

		sector, err := i.getStockInfo(line[4], stock.StockInfoSector)
		if err != nil {
			return err
		}

		industry, err := i.getStockInfo(line[5], stock.StockInfoIndustry)
		if err != nil {
			return err
		}

		ss = append(ss, stock.NewStock(m, e, line[0], line[2], t, sector, industry))
	}

	return i.purchaseService.SaveAllStocks(ss)
}

func (i *ImportStock) getStockInfo(value string, t stock.InfoType) (*stock.Info, error) {
	if stkInfo, ok := i.stockInfos[value]; ok {
		return stkInfo, nil
	}

	stkInfo, err := i.purchaseService.FindStockInfoByValue(value)
	if err != nil {
		if err != mm.ErrNotFound {
			return nil, err
		}

		stkInfo = stock.NewStockInfo(value, t)

		err = i.purchaseService.SaveStockInfo(stkInfo)
		if err != nil {
			return nil, err
		}

		i.stockInfos[value] = stkInfo
	}

	return stkInfo, nil
}
