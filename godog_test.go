package main

import (
	"flag"
	"os"
	"testing"
	"time"

	"log"

	"github.com/DATA-DOG/godog"
	"github.com/jmoiron/sqlx"

	"github.com/dohernandez/market-manager/features/bootstrap"
	"github.com/dohernandez/market-manager/pkg/config"
)

type deferredCall func()

var (
	runGoDogTests bool
	stopOnFailure bool
)

var deferredCalls []deferredCall

func init() {
	flag.BoolVar(&runGoDogTests, "godog", false, "Set this flag is you want to run godog BDD tests")
	flag.BoolVar(&stopOnFailure, "stop-on-failure", false, "Stop processing on first failed scenario.. Flag is passed to godog")
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
	bootstrap.StockCommandContext(s, db, conf.Import.StocksPath)

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
	})

	if st := m.Run(); st > status {
		status = st
	}

	for _, deferredCall := range deferredCalls {
		deferredCall()
	}

	os.Exit(status)
}
