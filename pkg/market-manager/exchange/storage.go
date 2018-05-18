package exchange

type Finder interface {
	FindBySymbol(symbol string) (*Exchange, error)
}
