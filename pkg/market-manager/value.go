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
	Euro           Currency = "€"
	Dollar                  = "$"
	CanadianDollar          = "C$"
)

var (
	exchangeCurrency map[string]Currency
)

func init() {
	exchangeCurrency = map[string]Currency{
		"NASDAQ": Dollar,
		"NYSE":   Dollar,
		"BME":    Euro,
		"FRA":    Euro,
		"BIT":    Euro,
	}
}

func (v *Value) Increase(a Value) Value {
	nv := Value{
		Currency: v.Currency,
	}
	nv.Amount = v.Amount + a.Amount

	return nv
}

func (v *Value) Decrease(a Value) Value {
	nv := Value{
		Currency: v.Currency,
	}
	nv.Amount = v.Amount - a.Amount

	return nv
}

// Compare to values
// 1 - gt than
// 0 - eq
// -1 - lt than
func (v *Value) Compare(a Value) int {
	if v.Amount > a.Amount {
		return 1
	}

	if v.Amount < a.Amount {
		return -1
	}

	return 0
}

func valueFromString(s string) Value {
	v, _ := strconv.ParseFloat(s, 64)

	return Value{Amount: v}
}

func ValueEuroFromString(s string) Value {
	a := valueFromString(s)
	a.Currency = Euro

	return a
}

func ValueDollarFromString(s string) Value {
	a := valueFromString(s)
	a.Currency = Dollar

	return a
}

func ValueFromStringAndExchange(s, e string) Value {
	a := valueFromString(s)
	a.Currency = ExchangeCurrency(e)

	return a
}

func ExchangeCurrency(e string) Currency {
	ec, ok := exchangeCurrency[e]
	if !ok {
		ec = Euro
	}

	return ec
}
