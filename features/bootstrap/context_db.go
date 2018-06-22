package bootstrap

import (
	"fmt"
	"log"
	"strings"

	"strconv"
	"time"

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
			"wallet_item",
			"operation",
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
	s.Step(`^that the following transfers are stored:$`, dbc.thatTheFollowingTransfersAreStored)
	s.Step(`^that the following stocks info are stored:$`, dbc.thatTheFollowingStocksInfoAreStored)
	s.Step(`^that the following stocks are stored:$`, dbc.thatTheFollowingStocksAreStored)

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

	for _, row := range data.Rows[1:] {
		ID, _ := uuid.FromString(row.Cells[0].Value)

		args := []interface{}{
			ID,
		}
		for k, cell := range row.Cells[1:] {
			if data.Rows[0].Cells[k+1].Value == "date" || data.Rows[0].Cells[k+1].Value == "last_price_update" ||
				data.Rows[0].Cells[k+1].Value == "high_low_52_week_update" {
				args = append(args, parseDateString(cell.Value))

				continue
			}

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
func (c *DBContext) thatTheFollowingBankAccountsAreStored(bAccounts *gherkin.DataTable) error {
	return c.runStoreData("bank_account", bAccounts)
}

// thatTheFollowingWalletBankAccountsAreStored inserts into the db table wallet_bank_account the values from the DataTable.
func (c *DBContext) thatTheFollowingWalletBankAccountsAreStored(wbAccounts *gherkin.DataTable) error {
	return c.runStoreData("wallet_bank_account", wbAccounts)
}

// thatTheFollowingTransfersAreStored inserts into the db table transfer the values from the DataTable.
func (c *DBContext) thatTheFollowingTransfersAreStored(transfers *gherkin.DataTable) error {
	for _, row := range transfers.Rows[1:] {
		for k, cell := range row.Cells {
			if transfers.Rows[0].Cells[k].Value == "amount" {
				a, err := parsePriceString(cell.Value)
				if err != nil {
					return err
				}

				cell.Value = strconv.FormatFloat(a, 'E', -1, 64)
			}
		}
	}

	return c.runStoreData("transfer", transfers)
}

// thatTheFollowingStocksAreStored inserts into the db table stock_info the values from the DataTable.
func (c *DBContext) thatTheFollowingStocksInfoAreStored(stocksInfo *gherkin.DataTable) error {
	return c.runStoreData("stock_info", stocksInfo)
}

// thatTheFollowingStocksAreStored inserts into the db table stock the values from the DataTable.
func (c *DBContext) thatTheFollowingStocksAreStored(stocks *gherkin.DataTable) error {
	for _, row := range stocks.Rows[1:] {
		for k, cell := range row.Cells {
			if stocks.Rows[0].Cells[k].Value == "exchange_symbol" {
				var id string
				err := c.db.Get(&id, `SELECT id  FROM exchange WHERE symbol like $1`, cell.Value)
				if err != nil {
					return err
				}

				cell.Value = id

				continue
			}

			if stocks.Rows[0].Cells[k].Value == "market_name" {
				var id string
				err := c.db.Get(&id, `SELECT id  FROM market WHERE name like $1`, cell.Value)
				if err != nil {
					return err
				}

				cell.Value = id
			}
		}
	}

	for _, cell := range stocks.Rows[0].Cells {
		if cell.Value == "exchange_symbol" {
			cell.Value = "exchange_id"

			continue
		}

		if cell.Value == "market_name" {
			cell.Value = "market_id"
		}
	}

	return c.runStoreData("stock", stocks)
}

// parseDateString - parse a potentially partial date string to Time
func parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2/1/2006", dt)

	return t
}

// parsePriceString - parse a potentially float string to float64
func parsePriceString(price string) (float64, error) {
	price = strings.Replace(price, ".", "", 1)
	price = strings.Replace(price, ",", ".", 1)

	return strconv.ParseFloat(price, 64)
}
