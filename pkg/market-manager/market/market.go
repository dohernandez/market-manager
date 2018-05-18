package market

import "github.com/satori/go.uuid"

const (
	Stock          = "stock"
	Cryptocurrency = "cryptocurrency"
)

// Market represents market struct
type Market struct {
	ID          uuid.UUID
	Name        string
	DisplayName string `db:"display_name"`
}

func NewMarket(name, displayName string) *Market {
	return &Market{
		ID:          uuid.NewV4(),
		Name:        name,
		DisplayName: displayName,
	}
}
