package account

type (
	Service struct {
		accountPersister Persister
	}
)

func NewService(accountPersister Persister) *Service {
	return &Service{
		accountPersister: accountPersister,
	}
}

func (s *Service) SaveAll(as []*Account) error {
	return s.accountPersister.PersistAll(as)
}
