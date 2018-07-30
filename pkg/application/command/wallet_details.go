package command

type WalletDetails struct {
	Wallet string

	Sells map[string]int
	Buys  map[string]int
}
