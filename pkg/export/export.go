package export

import (
	"fmt"
	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	Export interface {
		Export() error
	}

	OrderBy string
	SortBy  string
	GroupBy string

	Sorting struct {
		By    SortBy
		Order OrderBy
	}
)

const (
	Descending OrderBy = "desc"
	Ascending  OrderBy = "asc"
)

func PrintValue(value mm.Value, precision int) string {
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

func PrintDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format("02 Jan 2006")
}

func PrintDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format("02 Jan 06 15:04")
}
