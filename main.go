package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/application/cmd"
	"github.com/dohernandez/market-manager/pkg/application/config"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
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
	baseCMD := cmd.NewBaseCMD(context.TODO(), envConfig)
	baseExportCMD := &cmd.BaseExportCMD{}
	baseImportCMD := &cmd.BaseImportCMD{}

	serverCMD := cmd.NewHTTPCMD(baseCMD)

	migrateCMD := cmd.NewMigrateCMD(baseCMD)
	purchaseCMD := cmd.NewPurchaseCMD(baseCMD, baseImportCMD, baseExportCMD)
	bankingCMD := cmd.NewBankingCMD(baseCMD, baseImportCMD)
	accountCMD := cmd.NewAccountCMD(baseCMD, baseImportCMD, baseExportCMD)
	apiCMD := cmd.NewApiCMD(baseCMD)
	//
	//schedulerCMD := command.NewSchedulerCMD(purchaseCMD)

	app.Commands = []cli.Command{
		{
			Name:   "http",
			Usage:  "Start REST API service",
			Action: serverCMD.Run,
		},
		//	{
		//		Name:   "scheduler",
		//		Usage:  "Start Scheduler service",
		//		Action: schedulerCommand.Run,
		//	},
		{
			Name:      "migrate",
			Aliases:   []string{"m"},
			Usage:     "Run database migrations to the specific version",
			Action:    migrateCMD.Run,
			ArgsUsage: "",
			Subcommands: []cli.Command{
				{
					Name:      "up",
					Aliases:   []string{"u"},
					Usage:     "Up the database migrations",
					Action:    migrateCMD.Up,
					ArgsUsage: "",
				},
				{
					Name:      "down",
					Aliases:   []string{"d"},
					Usage:     "Down the database migrations",
					Action:    migrateCMD.Down,
					ArgsUsage: "",
				},
			},
		},
		{
			Name:    "purchase",
			Aliases: []string{"p"},
			Usage:   "Add/Update purchase values (markets, exchanges, stocks, cryptos)",
			Subcommands: []cli.Command{
				{
					Name:    "import",
					Aliases: []string{"i"},
					Usage:   "Import from csv file",
					Subcommands: []cli.Command{
						{
							Name:      "stock",
							Aliases:   []string{"q"},
							Usage:     "Import market stock from csv file",
							Action:    purchaseCMD.ImportStock,
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
							Usage:     "Import market stock dividend from csv file",
							Action:    purchaseCMD.ImportDividend,
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
							},
						},
					},
				},
				{
					Name:    "update",
					Aliases: []string{"u"},
					Usage:   "Update data",
					Subcommands: []cli.Command{
						{
							Name:      "price",
							Aliases:   []string{"p"},
							Usage:     "Update stock price value based on the yahoo/google api",
							Action:    purchaseCMD.UpdatePrice,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "stock, s",
									Usage: "Stock symbol(tricker) to update price",
								},
							},
						}, {
							Name:      "highLow52week",
							Aliases:   []string{"phl52w"},
							Usage:     "Update high - Low 52 week stock price value based on the yahoo/google api",
							Action:    purchaseCMD.UpdatePrice52week,
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
							Action:    purchaseCMD.Dividend,
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
					Name:    "export",
					Aliases: []string{"e"},
					Usage:   "Export to csv file",
					Subcommands: []cli.Command{
						{
							Name:      "stocks",
							Aliases:   []string{"s"},
							Action:    purchaseCMD.ExportStocks,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to export",
								},
								cli.StringFlag{
									Name:  "exchange, e",
									Usage: "filter by exchange",
								},
								cli.StringFlag{
									Name:  "stock, s",
									Usage: "filter by stock",
								},
								cli.StringFlag{
									Name:  "sort",
									Usage: "Sort by (name, dyield, exdate) Default by name",
								},
								cli.StringFlag{
									Name:  "order",
									Usage: "Order (desc, asc) Default by desc",
								},
								cli.StringFlag{
									Name:  "group",
									Usage: "Group by (exdate) None by default",
								},
							},
						},
						{
							Name:      "stocksDividends",
							Aliases:   []string{"sd"},
							Action:    purchaseCMD.ExportStocksWithDividend,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to export",
								},
								cli.StringFlag{
									Name:  "year, y",
									Usage: "filter by year. Default current year",
								},
								cli.StringFlag{
									Name:  "month, m",
									Usage: "filter by month. Default current month",
								},
								cli.StringFlag{
									Name:  "sort",
									Usage: "Sort by (dyield, exdate) Default by dyield",
								},
								cli.StringFlag{
									Name:  "order",
									Usage: "Order (desc, asc) Default by desc",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:    "account",
			Aliases: []string{"a"},
			Usage:   "Add/Update account values (operations)",
			Subcommands: []cli.Command{
				{
					Name:    "import",
					Aliases: []string{"i"},
					Usage:   "Import from csv file",
					Subcommands: []cli.Command{
						{
							Name:      "wallet",
							Aliases:   []string{"w"},
							Action:    accountCMD.ImportWallet,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to import",
								},
								cli.StringFlag{
									Name:  "wallet, w",
									Usage: "Wallet name",
								},
							},
						},
						{
							Name:      "operation",
							Aliases:   []string{"o"},
							Action:    accountCMD.ImportOperation,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to import",
								},
								cli.StringFlag{
									Name:  "wallet, w",
									Usage: "Wallet name",
								},
							},
						},
					},
				},
				{
					Name:    "export",
					Aliases: []string{"e"},
					Usage:   "Export to csv file",
					Subcommands: []cli.Command{
						{
							Name:      "walletItems",
							Aliases:   []string{"wi"},
							Action:    accountCMD.ExportWalletItems,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to export",
								},
								cli.StringFlag{
									Name:  "wallet, w",
									Usage: "Wallet name",
								},
								cli.StringFlag{
									Name:  "stock, s",
									Usage: "stock symbol",
								},
								cli.StringFlag{
									Name:  "sells, ss",
									Usage: "stock symbol to sell and amount separate by comma. Example epd:10",
								},
								cli.StringFlag{
									Name:  "buys, sb",
									Usage: "stock symbol to buy and amount separate by comma. Example epd:10",
								},
								cli.StringFlag{
									Name:  "sort",
									Usage: "Sort by (stock, invested) Default by stock",
								},
								cli.StringFlag{
									Name:  "order",
									Usage: "Order (desc, asc) Default by desc",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:    "banking",
			Aliases: []string{"a"},
			Usage:   "Add/Update banking values (transfers)",
			Subcommands: []cli.Command{
				{
					Name:    "import",
					Aliases: []string{"i"},
					Usage:   "Import from csv file",
					Subcommands: []cli.Command{
						{
							Name:      "transfer",
							Aliases:   []string{"t"},
							Action:    bankingCMD.ImportTransfer,
							ArgsUsage: "",
							Flags: []cli.Flag{
								cli.StringFlag{
									Name:  "file, f",
									Usage: "csv file to import",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:   "api",
			Usage:  "Api test",
			Action: apiCMD.Run,
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
