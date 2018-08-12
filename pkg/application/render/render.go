package render

import (
	"time"

	"github.com/satori/go.uuid"

	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

type (
	Render interface {
		Render(output interface{})
	}

	StockOutput struct {
		ID                uuid.UUID
		Stock             string
		Market            string
		Symbol            string
		Value             mm.Value
		High52Week        mm.Value
		Low52Week         mm.Value
		BuyUnder          mm.Value
		Dividend          mm.Value
		DividendRetention mm.Value
		DYield            float64
		DividendStatus    dividend.Status
		EPS               float64
		ExDate            time.Time
		Change            mm.Value
		UpdatedAt         time.Time
		HV52Week          float64
		HV20Day           float64

		PriceWithHighLow int
	}

	TradeOutput struct {
		ID     uuid.UUID
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

	WalletOutput struct {
		Capital                mm.Value
		Invested               mm.Value
		Funds                  mm.Value
		FreeMargin             mm.Value
		NetCapital             mm.Value
		NetBenefits            mm.Value
		PercentageBenefits     float64
		DividendPayed          mm.Value
		DividendMonthProjected mm.Value
		DividendMonthYield     float64
		DividendYearProjected  mm.Value
		DividendYearYield      float64
		Connection             mm.Value
		Interest               mm.Value
		Commission             mm.Value
	}

	WalletDetailsOutput struct {
		WalletOutput       WalletOutput
		WalletStockOutputs []*WalletStockOutput
	}
)
