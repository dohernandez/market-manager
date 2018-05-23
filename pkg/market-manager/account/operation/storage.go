package operation

type Persister interface {
	PersistAll(os []*Operation) error
}
