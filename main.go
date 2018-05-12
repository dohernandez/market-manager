package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/config"
	"github.com/dohernandez/market-manager/pkg/logger"
)

// Build version. Sent as a linker flag in Makefile
var version string

var binaryName = "market-manager-service"

func main() {
	app := cli.NewApp()
	app.Version = version
	app.Name = binaryName
	app.Usage = "The (first) awesome service for operation manager!"
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
	//baseCommand := command.NewBaseCommand(context.TODO(), envConfig)
	//serverCommand := command.NewHTTPCommand(baseCommand)
	//
	//app.Commands = []cli.Command{
	//	{
	//		Name:   "http",
	//		Usage:  "Start REST API service",
	//		Action: serverCommand.Run,
	//	},
	//}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
