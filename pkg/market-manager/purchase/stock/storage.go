package stock

import "github.com/satori/go.uuid"

type (
	Finder interface {
		FindAll() ([]*Stock, error)
		FindByID(ID uuid.UUID) (*Stock, error)
		FindBySymbol(symbol string) (*Stock, error)
		FindByName(name string) (*Stock, error)
	}

	Persister interface {
		PersistAll(ss []*Stock) error
		UpdatePrice(s *Stock) error
		UpdateDividendYield(s *Stock) error
	}
)
