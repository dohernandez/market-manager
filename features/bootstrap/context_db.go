package bootstrap

import (
	"log"

	"fmt"

	"strings"

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
			"transfer",
			"wallet_bank_account",
			"wallet",
			"stock",
			"stock_info",
			"import",
		},
	}

	s.BeforeScenario(func(i interface{}) {
		dbc.cleanUpDB()
	})

	s.Step(`^that the following wallets are stored:$`, dbc.thatTheFollowingWalletsAreStored)

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

func (c *DBContext) thatTheFollowingWalletsAreStored(wallets *gherkin.DataTable) error {
	var tColumns []string
	for _, cell := range wallets.Rows[0].Cells {
		tColumns = append(tColumns, cell.Value)
	}

	if len(tColumns) == 0 {
		return fmt.Errorf("there is no column defined INSERT INTO wallet")
	}

	query := fmt.Sprintf(`INSERT INTO wallet (%s) VALUES ($1, $2, $3)`, strings.Join(tColumns, ", "))

	for _, row := range wallets.Rows[1:] {
		ID, _ := uuid.FromString(row.Cells[0].Value)

		_, err := c.db.Exec(
			query,
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

//
//func (c *DBContext) thatTheFollowingMarketsAreStored(markets *gherkin.DataTable) error {
//	for _, row := range markets.Rows[1:] {
//		ID, _ := uuid.FromString(row.Cells[0].Value)
//
//		_, err := c.db.Exec(
//			`INSERT INTO market (id, name, display_name) VALUES ($1, $2, $3)`,
//			ID,
//			row.Cells[1].Value,
//			row.Cells[2].Value,
//		)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
