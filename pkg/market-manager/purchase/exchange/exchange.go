package exchange

import (
	"github.com/satori/go.uuid"
)

// Exchange represents exchange struct
type Exchange struct {
	ID     uuid.UUID
	Name   string
	Symbol string
}

func NewExchange(name, symbol string) *Exchange {
	return &Exchange{
		ID:     uuid.NewV4(),
		Name:   name,
		Symbol: symbol,
	}
}
