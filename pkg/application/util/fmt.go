package util

import (
	"fmt"
	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager"
)

func SPrintValue(value mm.Value, precision int) string {
	if value.Amount == 0 {
		return ""
	}

	if value.Currency == mm.Dollar {
		if value.Amount > 0 {
			return fmt.Sprintf("%s%.*f", value.Currency, precision, value.Amount)
		}

		return fmt.Sprintf("-%s%.*f", value.Currency, precision, value.Amount*-1)
	}

	return fmt.Sprintf("%.*f %s", precision, value.Amount, value.Currency)
}

func SPrintDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format("02 Jan 2006")
}

func SPrintDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format("02 Jan 06 15:04")
}
