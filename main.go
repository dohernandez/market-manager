package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/command"
	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// Build version. Sent as a linker flag in Makefile
var version string

var binaryName = "market-manager"

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Name = binaryName
	app.Usage = "The (first) awesome service for manage market!"
	app.UsageText = fmt.Sprintf("%s command [arguments]", binaryName)

	// Init envConfig
	envConfig, err := config.LoadEnv()
	if err != nil {
		logrus.WithError(err).Fatal("The envConfig could not be loaded.")
	}

	// Init logger
	log := logrus.StandardLogger()
	level, err := logrus.ParseLevel(envConfig.LogLevel)
	if err != nil {
		log.WithError(err).Fatal("Parsing LogLevel failed.")
	}
	log.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{FieldMap: logrus.FieldMap{
		logrus.FieldKeyTime: "@timestamp",
		logrus.FieldKeyMsg:  "message",
	}})

	logger.SetLogger(log)

	// Init command handlers
	// TODO: Real ctx should be passed here
	baseCommand := command.NewBaseCommand(context.TODO(), envConfig)
	serverCommand := command.NewHTTPCommand(baseCommand)
	migrateCommand := command.NewMigrateCommand(baseCommand)

	app.Commands = []cli.Command{
		{
			Name:   "http",
			Usage:  "Start REST API service",
			Action: serverCommand.Run,
		},
		{
			Name:      "migrate",
			Aliases:   []string{"m"},
			Usage:     "Run database migrations to the specific version",
			Action:    migrateCommand.Run,
			ArgsUsage: "",
			Subcommands: []cli.Command{
				{
					Name:      "up",
					Aliases:   []string{"u"},
					Usage:     "Up the database migrations",
					Action:    migrateCommand.Up,
					ArgsUsage: "",
				},
				{
					Name:      "down",
					Aliases:   []string{"d"},
					Usage:     "Down the database migrations",
					Action:    migrateCommand.Down,
					ArgsUsage: "",
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
