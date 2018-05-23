package import_market

import (
	"context"
	"io"

	"github.com/dohernandez/market-manager/pkg/import"
	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	ImportStock struct {
		ctx             context.Context
		reader          _import.Reader
		purchaseService *purchase.Service
	}
)

func NewImportStock(
	ctx context.Context,
	reader _import.Reader,
	purchaseService *purchase.Service,
) *ImportStock {
	return &ImportStock{
		ctx:             ctx,
		reader:          reader,
		purchaseService: purchaseService,
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

		ss = append(ss, stock.NewStock(m, e, line[0], line[2]))
	}

	return i.purchaseService.SaveAllStocks(ss)
}
