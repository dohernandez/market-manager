package cmd

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
	// BaseCMD hold common command properties
	BaseCMD struct {
		ctx    context.Context
		config *config.Specification
		cache  *cache.Cache
	}
)

// NewBaseCMD creates a structure with common shared properties of the commands
func NewBaseCMD(ctx context.Context, config *config.Specification) *BaseCMD {
	return &BaseCMD{
		ctx:    ctx,
		config: config,
		cache:  cache.New(time.Hour*2, time.Hour*10),
	}
}

func (cmd *BaseCMD) initDatabaseConnection() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cmd.config.Database.DSN)
	if err != nil {
		return nil, errors.Wrap(err, "Connecting to postgres")
	}

	return db, nil
}

func (cmd *BaseCMD) Container(db *sqlx.DB) *app.Container {
	return app.NewContainer(cmd.ctx, db, cmd.config, cmd.cache)
}
