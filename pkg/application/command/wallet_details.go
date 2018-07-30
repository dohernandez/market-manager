package command

type (
	Commission struct {
		Commission struct {
			Base struct {
				Amount   float64
				Currency string
			}
			Extra struct {
				Amount   float64
				Currency string
				Apply    string
			}
			Maximum struct {
				Amount   float64
				Currency string
			}
		}

		ChangeCommission struct {
			Amount   float64
			Currency string
		}
	}

	WalletDetails struct {
		Wallet string

		Sells map[string]int
		Buys  map[string]int

		Commissions map[string]Commission
	}
)
