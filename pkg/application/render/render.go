package render

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
)

type (
	Render interface {
		Render(output interface{})
	}

	StockOutput struct {
		ID         uuid.UUID
		Stock      string
		Market     string
		Symbol     string
		Value      mm.Value
		High52Week mm.Value
		Low52Week  mm.Value
		BuyUnder   mm.Value
		DYield     float64
		EPS        float64
		ExDate     time.Time
		Change     mm.Value
		UpdatedAt  time.Time

		PriceWithHighLow int
	}
)
