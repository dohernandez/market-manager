package dividend

import (
	"time"

	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	Status string

	StockDividend struct {
		ExDate      time.Time
		PaymentDate time.Time
		RecordDate  time.Time

		Status Status

		Amount mm.Value

		ChangeFromPrev     float64
		ChangeFromPrevYear float64
		Prior12MonthsYield float64
	}
)

const (
	Projected Status = "projected"
	Announced Status = "announced"
	Payed     Status = "payed"
)
