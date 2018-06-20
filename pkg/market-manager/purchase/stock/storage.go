package stock

import "github.com/satori/go.uuid"

type (
	Finder interface {
		FindAll() ([]*Stock, error)
		FindByID(ID uuid.UUID) (*Stock, error)
		FindBySymbol(symbol string) (*Stock, error)
		FindByName(name string) (*Stock, error)
		FindAllByExchanges(exchanges []string) ([]*Stock, error)
		FindAllByDividendAnnounceProjectYearAndMonth(year, month int) ([]*Stock, error)
	}

	Persister interface {
		PersistAll(ss []*Stock) error
		UpdatePrice(s *Stock) error
		UpdateDividendYield(s *Stock) error
		UpdateHighLow52WeekPrice(s *Stock) error
	}

	InfoFinder interface {
		FindByName(name string) (*Info, error)
	}

	InfoPersister interface {
		Persist(i *Info) error
	}
)
