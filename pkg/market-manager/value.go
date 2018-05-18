package mm

type Value struct {
	Amount   int
	Currency Currency
}

// Type holds different compensation type definitions
type Currency string

// Possible currency
const (
	Euro   Currency = "€"
	Dollar          = "$"
)
