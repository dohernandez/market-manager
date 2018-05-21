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
	importCommand := command.NewImportCommand(baseCommand)
	stocksCommand := command.NewStocksCommand(baseCommand)
	apiCommand := command.NewApiCommand(baseCommand)

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
		{
			Name:    "stocks",
			Aliases: []string{"s"},
			Usage:   "Add/Update stock values",
			Subcommands: []cli.Command{
				{
					Name:    "import",
					Aliases: []string{"i"},
					Usage:   "Import stock from csv file",
					Subcommands: []cli.Command{
						{
							Name:      "quote",
							Aliases:   []string{"s"},
							Usage:     "Import stock from csv file",
							Action:    importCommand.Quote,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to import",
								},
							},
						},
						{
							Name:      "dividend",
							Aliases:   []string{"s"},
							Usage:     "Import stock from csv file",
							Action:    importCommand.Dividend,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to import",
								},
								cli.StringFlag{
									Name:  "stock, s",
									Usage: "Stock symbol (tricker) to update dividend.",
								},
								cli.StringFlag{
									Name:  "status, st",
									Usage: "Dividend status [payed, projected]. Default: payed",
								},
							},
						},
					},
				},
				{
					Name:      "price",
					Aliases:   []string{"p"},
					Usage:     "Update stock price value based on the yahoo/google api",
					Action:    stocksCommand.Price,
					ArgsUsage: "",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "stock, s",
							Usage: "Stock symbol(tricker) to update price",
						},
					},
				},
				{
					Name:      "dividend",
					Aliases:   []string{"d"},
					Usage:     "Update stock dividend value based on the yahoo/iextrading api",
					Action:    stocksCommand.Dividend,
					ArgsUsage: "",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "stock, s",
							Usage: "Stock symbol(tricker) to update price",
						},
					},
				},
			},
		},
		{
			Name:   "api",
			Usage:  "Api test",
			Action: apiCommand.Run,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
