package account

type Persister interface {
	PersistAll(as []*Account) error
}
