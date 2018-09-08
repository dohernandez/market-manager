package command

type (
	WalletDateDetails struct {
		Wallet        string
		Date          string
		TransferPath  string
		OperationPath string
		Excludes      []string
	}
)
