package command

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application"
	"github.com/dohernandez/market-manager/pkg/application/config"
)

type (
	// BaseCommand hold common command properties
	BaseCommand struct {
		ctx    context.Context
		config *config.Specification
		cache  *cache.Cache
	}
)

// NewBaseCommand creates a structure with common shared properties of the commands
func NewBaseCommand(ctx context.Context, config *config.Specification) *BaseCommand {
	return &BaseCommand{
		ctx:    ctx,
		config: config,
		cache:  cache.New(time.Hour*2, time.Hour*10),
	}
}

func (cmd *BaseCommand) initDatabaseConnection() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cmd.config.Database.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "Connecting to postgres")
	}

	return db, nil
}

func (cmd *BaseCommand) Container(db *sqlx.DB) *app.Container {
	return app.NewContainer(cmd.ctx, db, cmd.config, cmd.cache)
}