package render

import (
	"os"
	"sort"
	"text/tabwriter"

	"github.com/fatih/color"

	"context"
)

type (
	screenWalletDateDetails struct {
		ctx context.Context
		screenWalletDetails
	}
)

func NewWalletDateDetails(ctx context.Context) *screenWalletDateDetails {
	return &screenWalletDateDetails{
		ctx: ctx,
	}
}

func (s *screenWalletDateDetails) Render(output interface{}) {
	sOutput := output.(*OutputScreenWalletDetails)

	walletStockOutputs := sOutput.WalletDetails.WalletStockOutputs
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
	s.renderStocks(tw, walletStockOutputs, precision)
	noColor(tw, "")

	tw.Flush()
}
