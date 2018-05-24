package bank

type (
	Finder interface {
		FindByAlias(alias string) (*Account, error)
	}
)
