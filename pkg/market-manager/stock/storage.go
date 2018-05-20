package stock

type (
	Finder interface {
		FindAll() ([]*Stock, error)
		FindBySymbol(symbol string) (*Stock, error)
	}

	Persister interface {
		PersistAll(ss []*Stock) error
		UpdatePrice(s *Stock) error
	}
)
