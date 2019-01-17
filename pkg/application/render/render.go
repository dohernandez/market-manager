package render

import (
	"time"

	uuid "github.com/satori/go.uuid"

	mm "github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	Render interface {
		Render(output interface{})
	}

	StockOutput struct {
		ID                  uuid.UUID
		Stock               string
		Market              string
		Symbol              string
		Value               mm.Value
		High52Week          mm.Value
		Low52Week           mm.Value
		BuyUnder            mm.Value
		Dividend            mm.Value
		DividendRetention   mm.Value
		PercentageRetention float64
		DYield              float64
		DividendStatus      dividend.Status
		EPS                 float64
		ExDate              time.Time
		Change              mm.Value
		UpdatedAt           time.Time
		HV52Week            float64
		HV20Day             float64
		PER                 float64

		PriceWithHighLow int
	}

	TradeOutput struct {
		ID     uuid.UUID
		Number int
		Stock  string
		Market string
		Symbol string

		Enter struct {
			Amount float64
			Kurs   mm.Value
			Total  mm.Value
		}

		Position struct {
			Amount   float64
			Dividend mm.Value
			Capital  mm.Value
		}

		Exit struct {
			Amount float64
			Kurs   mm.Value
			Total  mm.Value
		}

		Net               mm.Value
		BenefitPercentage float64

		IsProfitable bool
	}

	WalletStockOutput struct {
		StockOutput
		Amount             int
		Capital            mm.Value
		Invested           mm.Value
		DividendPayed      mm.Value
		DividendToPay      mm.Value
		PercentageWallet   float64
		Buys               mm.Value
		Sells              mm.Value
		NetBenefits        mm.Value
		PercentageBenefits float64
		Change             mm.Value
		WAPrice            mm.Value
		WADYield           float64
		Trades             []*TradeOutput
	}

	WalletDividendProjected struct {
		Month     string
		Projected mm.Value
		Yield     float64
	}

	WalletOutput struct {
		Capital                mm.Value
		Invested               mm.Value
		Funds                  mm.Value
		FreeMargin             mm.Value
		NetCapital             mm.Value
		NetBenefits            mm.Value
		PercentageBenefits     float64
		DividendPayed          mm.Value
		DividendPayedYield     float64
		DividendYearProjected  mm.Value
		DividendYearYield      float64
		DividendTotalProjected mm.Value
		DividendTotalYield     float64
		Connection             mm.Value
		Interest               mm.Value
		Commission             mm.Value

		DividendProjected []WalletDividendProjected
	}

	WalletDetailsOutput struct {
		WalletOutput       WalletOutput
		WalletStockOutputs []*WalletStockOutput
	}
)
