package bootstrap

import (
	"fmt"
	"log"
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
			"bank_account",
			"import",
		},
	}

	s.BeforeScenario(func(i interface{}) {
		dbc.cleanUpDB()
	})

	s.Step(`^that the following wallets are stored:$`, dbc.thatTheFollowingWalletsAreStored)
	s.Step(`^that the following bank accounts are stored:$`, dbc.thatTheFollowingBankAccountsAreStored)
	s.Step(`^that the following wallet bank accounts are stored:$`, dbc.thatTheFollowingWalletBankAccountsAreStored)

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

// thatTheFollowingWalletsAreStored inserts into the db table wallet the values from the DataTable.
func (c *DBContext) thatTheFollowingWalletsAreStored(wallets *gherkin.DataTable) error {
	return c.runStoreData("wallet", wallets)
}

// runStoreData inserts into the db table the values from the DataTable.
// The DataTable could expand dynamically as many column contain the table, resulting in some cases a DataTable
// with 3 columns or 5 columns depends on the background data require
func (c *DBContext) runStoreData(table string, data *gherkin.DataTable) error {
	var tColumns, qValue []string
	for i, cell := range data.Rows[0].Cells {
		tColumns = append(tColumns, cell.Value)
		qValue = append(qValue, fmt.Sprintf("$%d", i+1))
	}

	if len(tColumns) == 0 {
		return fmt.Errorf("there is no column defined INSERT INTO wallet")
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		table,
		strings.Join(tColumns, ", "),
		strings.Join(qValue, ", "),
	)

	fmt.Println(query)

	for _, row := range data.Rows[1:] {
		ID, _ := uuid.FromString(row.Cells[0].Value)

		args := []interface{}{
			ID,
		}
		for _, cell := range row.Cells[1:] {
			args = append(args, cell.Value)
		}

		_, err := c.db.Exec(query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

// thatTheFollowingBankAccountsAreStored inserts into the db table bank_account the values from the DataTable.
func (c *DBContext) thatTheFollowingBankAccountsAreStored(bAccount *gherkin.DataTable) error {
	return c.runStoreData("bank_account", bAccount)
}

// thatTheFollowingWalletBankAccountsAreStored inserts into the db table wallet_bank_account the values from the DataTable.
func (c *DBContext) thatTheFollowingWalletBankAccountsAreStored(wbAccount *gherkin.DataTable) error {
	return c.runStoreData("wallet_bank_account", wbAccount)
}
