package mm

import (
	"strconv"
)

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

func (v *Value) Increase(a Value) Value {
	nv := Value{}
	nv.Amount = v.Amount + a.Amount

	return nv
}

func (v *Value) Decrease(a Value) Value {
	nv := Value{}
	nv.Amount = v.Amount - a.Amount

	return nv
}

func ValueFromString(s string) Value {
	v, _ := strconv.ParseFloat(s, 64)

	return Value{Amount: v}
}
