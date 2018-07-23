package cmd

import (
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/application/export"
	exportAccount "github.com/dohernandez/market-manager/pkg/application/export/account"
)

type (
	Sorting struct {
	}

	// BaseExport ...
	BaseExport struct {
	}
)

func (e *BaseExport) sortingFromCliCtx(cliCtx *cli.Context) export.Sorting {
	sortBy := exportAccount.Stock
	orderBy := export.Descending

	if cliCtx.String("sort") != "" {
		sortBy = export.SortBy(cliCtx.String("sort"))
	}
	if cliCtx.String("order") != "" {
		orderBy = export.OrderBy(cliCtx.String("order"))
	}

	return export.Sorting{
		By:    sortBy,
		Order: orderBy,
	}
}

func (e *BaseExport) runExport(ex export.Export) error {
	return ex.Export()
}
