package command

import "github.com/dohernandez/market-manager/pkg/application/util"

type ListStocks struct {
	Exchange string

	GroupBy util.GroupBy
	Sorting util.Sorting
}
