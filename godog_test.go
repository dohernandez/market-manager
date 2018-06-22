package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/jmoiron/sqlx"

	"fmt"

	"github.com/dohernandez/market-manager/features/bootstrap"
	"github.com/dohernandez/market-manager/pkg/config"
)

type deferredCall func()

var deferredCalls []deferredCall

var (
	runGoDogTests bool
	stopOnFailure bool
	runWithTags   string
	runFeature    string
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
	flag.StringVar(&runFeature, "feature", "", "Optional feature to run. Filename without the extension .feature")

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
	bootstrap.RegisterCsvFileContext(s, "resources/dev/import")
	bootstrap.RegisterStockCommandContext(s, db)
	bootstrap.RegisterAccountCommandContext(s, db)
	bootstrap.RegisterBankingCommandContext(s, db)

	//ctx := context.TODO()

	return
}

func TestMain(m *testing.M) {
	if !runGoDogTests {
		os.Exit(0)
	}

	paths := []string{"features"}
	if runFeature != "" {
		paths = []string{fmt.Sprintf("features/%s.feature", runFeature)}
	}

	status := godog.RunWithOptions("MarketManager", func(s *godog.Suite) {
		FeatureContext(s)
	}, godog.Options{
		Format:        "pretty",
		Paths:         paths,
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
