package command

import "github.com/urfave/cli"

// BankingCommand ...
type BankingCommand struct {
	*BaseCommand
}

// NewBankingCommand constructs BankingCommand
func NewBankingCommand(baseCommand *BaseCommand) *BankingCommand {
	return &BankingCommand{
		BaseCommand: baseCommand,
	}
}

func (cmd *BankingCommand) Transfer(cliCtx *cli.Context) error {
	return nil
}
