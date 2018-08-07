package render

import (
	"os"
	"sort"
	"text/tabwriter"

	"github.com/fatih/color"

	"context"

	"fmt"

	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
)

type (
	OutputScreenWalletStockDetails struct {
		OutputScreenWalletDetails

		Stock string
	}

	screenStockWalletDetails struct {
		ctx context.Context
		screenWalletDetails
	}
)

func NewScreenWalletStockDetails(ctx context.Context) *screenStockWalletDetails {
	return &screenStockWalletDetails{
		ctx: ctx,
	}
}

func (s *screenStockWalletDetails) Render(output interface{}) {
	sOutput := output.(*OutputScreenWalletStockDetails)

	var walletStockOutputs []*WalletStockOutput

	for _, wStockOutputs := range sOutput.WalletDetails.WalletStockOutputs {
		if wStockOutputs.Symbol == sOutput.Stock {
			walletStockOutputs = append(walletStockOutputs, wStockOutputs)
		}
	}

	if len(walletStockOutputs) == 0 {
		logger.FromContext(s.ctx).Fatal("No stock found")
	}

	walletOutput := sOutput.WalletDetails.WalletOutput
	precision := sOutput.Precision

	switch sOutput.Sorting.By {
	case WalletInvested:
		sort.Sort(WalletItemsByInvested{walletStockOutputs})
	default:
		sort.Sort(WalletItemsByName{walletStockOutputs})
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	noColor := color.New(color.Reset).FprintlnFunc()

	noColor(tw, "")
	s.renderGeneral(tw, walletOutput, precision)
	noColor(tw, "")
	s.renderItemStocks(tw, walletStockOutputs, precision)
	noColor(tw, "")
	s.renderWalletDividendProjected(tw, walletOutput, precision)
	noColor(tw, "")
	s.renderStocksDividends(tw, walletStockOutputs, precision)
	noColor(tw, "")
	s.renderStocks(tw, walletStockOutputs, precision)
	noColor(tw, "")
	s.renderTrades(tw, walletStockOutputs, precision)
	noColor(tw, "")

	tw.Flush()
}

func (s *screenStockWalletDetails) renderTrades(tw *tabwriter.Writer, wStocks []*WalletStockOutput, precision int) {
	noColor := color.New(color.Reset).FprintlnFunc()
	noColor(tw, "# Trades")
	noColor(tw, "")

	header := color.New(color.FgWhite).FprintlnFunc()
	header(tw, "\t \t \t \t Enter\t \t \t Position\t \t \t Exit\t \t \t Benefit\t \t")
	header(tw, "#\t Stock\t Market\t Symbol\t Amount\t Value\t Total\t Amount\t Dividend\t Capital\t Amount\t Value\t Total\t Net\t %\t")

	profitable := color.New(color.FgGreen).FprintlnFunc()
	notProfitable := color.New(color.FgRed).FprintlnFunc()

	fmt.Printf("%s amount trades %d\n", wStocks[0].Stock, len(wStocks[0].Trades))
	for i, t := range wStocks[0].Trades {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s\t %0.f\t %s\t %s\t %.0f\t %s\t %s\t %.0f\t %s\t %s\t %s\t %s\t",
			i+1,
			t.Stock,
			t.Market,
			t.Symbol,
			t.Enter.Amount,
			util.SPrintValue(t.Enter.Kurs, precision),
			util.SPrintValue(t.Enter.Total, precision),
			t.Position.Amount,
			util.SPrintValue(t.Position.Dividend, precision),
			util.SPrintValue(t.Position.Capital, precision),
			t.Exit.Amount,
			util.SPrintValue(t.Exit.Kurs, precision),
			util.SPrintValue(t.Exit.Total, precision),
			util.SPrintValue(t.Net, precision),
			util.SPrintPercentage(t.BenefitPercentage, precision),
		)

		if t.IsProfitable {
			profitable(tw, str)

			continue
		}

		notProfitable(tw, str)
	}
}
