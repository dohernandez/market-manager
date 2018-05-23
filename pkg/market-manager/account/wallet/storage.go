package wallet

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type (
	Finder interface {
		FindByStock(s *stock.Stock) (*Item, error)
	}

	Persister interface {
		Persist(i *Item) error
		PersistAll(is []*Item) error
	}
)
