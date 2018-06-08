package dividend

import (
	"time"

	"github.com/satori/go.uuid"
)

type (
	Finder interface {
		FindAllFormStock(stockID uuid.UUID) ([]StockDividend, error)
		// Find the next dividend announce or projected for the stock
		FindNextFromStock(stockID uuid.UUID, dt time.Time) (StockDividend, error)
	}

	Persister interface {
		PersistAll(stockID uuid.UUID, ds []StockDividend) error
	}
)
