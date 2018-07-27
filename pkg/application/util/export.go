package util

type (
	OrderBy string
	SortBy  string
	GroupBy string

	Sorting struct {
		By    SortBy
		Order OrderBy
	}
)

const (
	OrderDescending OrderBy = "desc"
	OrderAscending  OrderBy = "asc"

	SortByStock  SortBy = "stock"
	SortInvested SortBy = "invested"
)
