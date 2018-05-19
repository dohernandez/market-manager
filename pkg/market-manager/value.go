package mm

type Value struct {
	Amount   float64
	Currency Currency
}

// Type holds different compensation type definitions
type Currency string

// Possible currency
const (
	Euro   Currency = "€"
	Dollar          = "$"
)
