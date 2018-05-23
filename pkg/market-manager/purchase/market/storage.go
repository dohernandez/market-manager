package market

type Finder interface {
	FindByName(name string) (*Market, error)
}
