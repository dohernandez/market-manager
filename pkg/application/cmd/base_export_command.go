package cmd

import (
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/export"
	exportAccount "github.com/dohernandez/market-manager/pkg/export/account"
)

type (
	Sorting struct {
	}

	// BaseExportCMD ...
	BaseExportCMD struct {
	}
)

func (e *BaseExportCMD) sortingFromCliCtx(cliCtx *cli.Context) export.Sorting {
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

func (e *BaseExportCMD) runExport(ex export.Export) error {
	return ex.Export()
}
