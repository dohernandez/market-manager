package bootstrap

import (
	"log"

	"fmt"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jmoiron/sqlx"
	"github.com/satori/go.uuid"
)

type DBContext struct {
	db     *sqlx.DB
	tables []string
}

// RegisterDBContext is the place to truncate database and run background givens
func RegisterDBContext(s *godog.Suite, db *sqlx.DB) *DBContext {
	dbc := DBContext{
		db: db,
		tables: []string{
			"stock",
			"stock_info",
			"exchange",
			"market",
			"import",
		},
	}

	s.BeforeScenario(func(i interface{}) {
		dbc.cleanUpDB()
	})

	s.Step(`^that the following exchanges are stored:$`, dbc.thatTheFollowingExchangesAreStored)
	s.Step(`^that the following markets are stored:$`, dbc.thatTheFollowingMarketsAreStored)

	return &dbc
}

func (c *DBContext) cleanUpDB() {
	for _, table := range c.tables {
		_, err := c.db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (c *DBContext) thatTheFollowingExchangesAreStored(exchanges *gherkin.DataTable) error {
	for _, row := range exchanges.Rows[1:] {
		ID, _ := uuid.FromString(row.Cells[0].Value)

		_, err := c.db.Exec(
			`INSERT INTO exchange (id, name, symbol) VALUES ($1, $2, $3)`,
			ID,
			row.Cells[1].Value,
			row.Cells[2].Value,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DBContext) thatTheFollowingMarketsAreStored(markets *gherkin.DataTable) error {
	for _, row := range markets.Rows[1:] {
		ID, _ := uuid.FromString(row.Cells[0].Value)

		_, err := c.db.Exec(
			`INSERT INTO market (id, name, display_name) VALUES ($1, $2, $3)`,
			ID,
			row.Cells[1].Value,
			row.Cells[2].Value,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
