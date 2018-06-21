package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"log"

	"github.com/DATA-DOG/godog"
	"github.com/jmoiron/sqlx"

	"github.com/DATA-DOG/godog/colors"

	"strings"

	"github.com/dohernandez/market-manager/features/bootstrap"
	"github.com/dohernandez/market-manager/pkg/config"
)

type deferredCall func()

var deferredCalls []deferredCall

var (
	runGoDogTests bool
	stopOnFailure bool
	runWithTags   string
)

func init() {
	flag.BoolVar(&runGoDogTests, "godog", false, "Set this flag is you want to run godog BDD tests")
	flag.BoolVar(&stopOnFailure, "stop-on-failure", false, "Stop processing on first failed scenario.. Flag is passed to godog")

	descTagsOption := "Filter scenarios by tags. Expression can be:\n" +
		strings.Repeat(" ", 4) + "- " + colors.Yellow(`"@test"`) + ": run all scenarios with wip tag\n" +
		strings.Repeat(" ", 4) + "- " + colors.Yellow(`"~@test"`) + ": exclude all scenarios with wip tag\n" +
		strings.Repeat(" ", 4) + "- " + colors.Yellow(`"@test && ~@new"`) + ": run wip scenarios, but exclude new\n" +
		strings.Repeat(" ", 4) + "- " + colors.Yellow(`"@test,@undone"`) + ": run wip or undone scenarios"

	flag.StringVar(&runWithTags, "tag", "", descTagsOption)

	flag.Parse()
}

func FeatureContext(s *godog.Suite) {
	conf, err := config.LoadEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Connect("postgres", conf.Database.DSN)
	if err != nil {
		log.Fatal(err)
	}

	bootstrap.RegisterDBContext(s, db)
	bootstrap.RegisterCommandContext(s)
	bootstrap.RegisterCsvFileContext(
		s,
		conf.Import.StocksPath,
		conf.Import.WalletsPath,
		conf.Import.TransfersPath,
	)
	bootstrap.RegisterStockCommandContext(s, db)
	bootstrap.RegisterAccountCommandContext(s, db)

	//ctx := context.TODO()

	return
}

func TestMain(m *testing.M) {
	if !runGoDogTests {
		os.Exit(0)
	}
	status := godog.RunWithOptions("MarketManager", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:        "pretty",
		Paths:         []string{"features"},
		Randomize:     time.Now().UTC().UnixNano(),
		StopOnFailure: stopOnFailure,
		Tags:          runWithTags,
	})

	if st := m.Run(); st > status {
		status = st
	}

	for _, deferredCall := range deferredCalls {
		deferredCall()
	}

	os.Exit(status)
}
