package export

type (
	Export interface {
		Export() error
	}

	OrderBy string
	SortBy  string

	Sorting struct {
		By    SortBy
		Order OrderBy
	}
)

const (
	Descending OrderBy = "desc"
	Ascending  OrderBy = "asc"
)
