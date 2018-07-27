package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"

	"github.com/dohernandez/market-manager/features/bootstrap"
	"github.com/dohernandez/market-manager/pkg/application/config"
)

type externalUrls struct {
	FinanceYahooBaseURL string `envconfig:"FINANCE_YAHOO_BASEURL"`
	Query1YahooBaseURL  string `envconfig:"QUERY1_YAHOO_BASEURL"`
}

var eUrls externalUrls

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

	if conf.Environment == "production" {
		log.Fatal(errors.New("can not run feature test in production"))
	}

	envconfig.Process("", &eUrls)

	db, err := sqlx.Connect("postgres", conf.Database.DSN)
	if err != nil {
		log.Fatal(err)
	}

	yahooFinanceWeb := bootstrap.RegisterWireMockExtension(s, eUrls.FinanceYahooBaseURL)
	query1FinanceWeb := bootstrap.RegisterWireMockExtension(s, eUrls.Query1YahooBaseURL)

	bootstrap.RegisterDBContext(s, db)
	bootstrap.RegisterCommandContext(s)
	bootstrap.RegisterCsvFileContext(s, "resources/dev/import")
	bootstrap.RegisterStockCommandContext(s, db, yahooFinanceWeb, query1FinanceWeb)
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
		paths = []string{fmt.Sprintf("features/%s", runFeature)}
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
