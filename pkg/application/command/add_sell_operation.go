package command

type AddSellOperation struct {
	Trade                 string
	Date                  string
	Wallet                string
	Stock                 string
	Price                 float64
	PriceChange           float64
	PriceChangeCommission float64
	Commission            float64
	Amount                int
	Value                 float64
}
