package command

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/container"
)

type (
	// BaseCommand hold common command properties
	BaseCommand struct {
		ctx    context.Context
		config *config.Specification
	}

	Sorting struct {
	}
)

// NewBaseCommand creates a structure with common shared properties of the commands
func NewBaseCommand(ctx context.Context, config *config.Specification) *BaseCommand {
	return &BaseCommand{
		ctx:    ctx,
		config: config,
	}
}

func (cmd *BaseCommand) initDatabaseConnection() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cmd.config.Database.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "Connecting to postgres")
	}

	return db, nil
}

func (cmd *BaseCommand) Container(db *sqlx.DB) *container.Container {
	return container.NewContainer(cmd.ctx, db, cmd.config)
}
