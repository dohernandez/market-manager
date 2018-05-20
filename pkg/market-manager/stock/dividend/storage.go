package dividend

import "github.com/satori/go.uuid"

type (
	Finder interface {
		FindAllFormStock(stockID uuid.UUID) ([]StockDividend, error)
	}

	Persister interface {
		PersistAll(stockID uuid.UUID, ds []StockDividend) error
	}
)
