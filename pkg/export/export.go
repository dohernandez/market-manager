package export

import (
	"fmt"

	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	Export interface {
		Export() error
	}

	OrderBy string
	SortBy  string

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
	if value.Currency == mm.Dollar {
		return fmt.Sprintf(" %s%.*f", value.Currency, precision, value.Amount)
	}

	return fmt.Sprintf("%.*f %s", precision, value.Amount, value.Currency)
}
