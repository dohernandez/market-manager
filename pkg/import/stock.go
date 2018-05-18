package _import

import (
	"context"
	"io"

	"github.com/dohernandez/market-manager/pkg/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/exchange"
	"github.com/dohernandez/market-manager/pkg/market-manager/market"
	"github.com/dohernandez/market-manager/pkg/market-manager/stock"
)

type (
	ImportStock struct {
		ctx            context.Context
		reader         Reader
		marketFinder   market.Finder
		exchangeFinder exchange.Finder
		stockPersister stock.Persister
	}
)

func NewImportStock(
	ctx context.Context,
	reader Reader,
	marketFinder market.Finder,
	exchangeFinder exchange.Finder,
	stockPersister stock.Persister,
) *ImportStock {
	return &ImportStock{
		ctx:            ctx,
		reader:         reader,
		marketFinder:   marketFinder,
		exchangeFinder: exchangeFinder,
		stockPersister: stockPersister,
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

		m, err := i.marketFinder.FindByName(market.Stock)
		if err != nil {
			return err
		}

		e, err := i.exchangeFinder.FindBySymbol(line[1])
		if err != nil {
			return err
		}

		ss = append(ss, stock.NewStock(m, e, line[0], line[2]))
	}

	return i.stockPersister.PersistAll(ss)
}
