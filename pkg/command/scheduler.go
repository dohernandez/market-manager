package command

import (
	"context"

	"github.com/jasonlvhit/gocron"
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/logger"
)

// SchedulerCommand ...
type SchedulerCommand struct {
	stockCommand *StocksCommand
}

// NewSchedulerCommand constructs SchedulerCommand
func NewSchedulerCommand(stockCommand *StocksCommand) *SchedulerCommand {
	return &SchedulerCommand{
		stockCommand: stockCommand,
	}
}

// Run ...
func (cmd *SchedulerCommand) Run(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	// Init scheduler
	logger.FromContext(ctx).Info("Starting schedulers")

	s := gocron.NewScheduler()
	s.Every(1).Day().At("8:20").Do(func() { cmd.stockCommand.Price(cliCtx) })
	<-s.Start()

	return nil
}
