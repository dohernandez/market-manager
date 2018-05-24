package transfer

type (
	Persister interface {
		PersistAll(ts []*Transfer) error
	}
)
