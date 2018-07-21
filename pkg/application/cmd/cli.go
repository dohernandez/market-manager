package cmd

import (
	"context"

	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/application/command"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
)

type CLI struct {
	*Base
}

func NewCLI(base *Base) *CLI {
	return &CLI{
		Base: base,
	}
}

func (cmd *CLI) UpdatePrice(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()

	bus := cmd.initCommandBus()

	if cliCtx.String("stock") == "" {
		_, err := bus.ExecuteContext(ctx, &command.UpdateAllStocksPrice{})
		if err != nil {
			return err
		}
	} else {
		_, err := bus.ExecuteContext(ctx, &command.UpdateOneStockPrice{Symbol: cliCtx.String("stock")})
		if err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Update finished")

	return nil
}
