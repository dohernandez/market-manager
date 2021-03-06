package cmd

import (
	"context"
	"strconv"

	"github.com/mattes/migrate"
	"github.com/sirupsen/logrus"
	// This is how database/sql operates
	_ "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
)

// Migrate ...
type Migrate struct {
	*Base
}

// NewMigrate constructs Migrate
func NewMigrate(base *Base) *Migrate {
	return &Migrate{
		Base: base,
	}
}

// Run runs the application database migrations
func (cmd *Migrate) Run(cliCtx *cli.Context) error {
	m, err := cmd.getMigrate()
	if err != nil {
		return err
	}

	if cliCtx.NArg() != 1 {
		logger.FromContext(context.TODO()).Error("Please specify the migration version: market-manager migrate [version]")
		return nil
	}

	version, err := strconv.Atoi(cliCtx.Args().Get(0))
	if err != nil {
		return errors.Wrap(err, "Failed parsing version argument")
	}

	return cmd.checkMigrationError(context.TODO(), m.Migrate(uint(version)))
}

// Up runs all the migrations
func (cmd *Migrate) Up(cliCtx *cli.Context) error {
	m, err := cmd.getMigrate()
	if err != nil {
		return err
	}

	return cmd.checkMigrationError(context.TODO(), m.Up())
}

// Down downs the migrations
func (cmd *Migrate) Down(cliCtx *cli.Context) error {
	m, err := cmd.getMigrate()
	if err != nil {
		return err
	}

	return cmd.checkMigrationError(context.TODO(), m.Down())
}

func (cmd *Migrate) getMigrate() (*migrate.Migrate, error) {
	logger.FromContext(context.TODO()).WithFields(logrus.Fields{
		"migrations_path": cmd.config.Database.MigrationsPath,
	}).Info("Initializing migration")

	m, err := migrate.New(
		cmd.config.Database.MigrationsPath,
		cmd.config.Database.DSN,
	)

	return m, errors.Wrap(err, "Initializing migrations failed")
}

func (cmd *Migrate) checkMigrationError(ctx context.Context, err error) error {
	if err == migrate.ErrNoChange {
		logger.FromContext(ctx).Info("No new migrations to run")
		return nil
	}

	if err != nil {
		return errors.Wrap(err, "Initializing migrations failed")
	}

	logger.FromContext(ctx).Info("Migration finished")

	return nil
}
