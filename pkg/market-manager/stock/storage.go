package stock

type Persister interface {
	PersistAll(ss []*Stock) error
}
