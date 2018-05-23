package dividend

import "time"

type (
	Status string

	StockDividend struct {
		ExDate      time.Time
		PaymentDate time.Time
		RecordDate  time.Time

		Status Status

		Amount float64

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
