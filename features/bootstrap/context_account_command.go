package bootstrap

import (
	"fmt"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jmoiron/sqlx"
)

type accountCommandContext struct {
	db          *sqlx.DB
	wallets     map[string]string
	operations  map[string]string
	walletItems map[string]string
}

func RegisterAccountCommandContext(s *godog.Suite, db *sqlx.DB) {
	acc := &accountCommandContext{
		db:          db,
		wallets:     map[string]string{},
		operations:  map[string]string{},
		walletItems: map[string]string{},
	}

	s.Step(`^following wallets should be stored:$`, acc.followingWalletsShouldBeStored)
	s.Step(`^the following wallets should have:$`, acc.theFollowingWalletsShouldHave)
	s.Step(`^the following operations should be stored:$`, acc.theFollowingOperationsShouldBeStored)
	s.Step(`^the following wallet items should be stored:$`, acc.theFollowingWalletItemsShouldBeStored)
}

func (c *accountCommandContext) followingWalletsShouldBeStored(wallets *gherkin.DataTable) error {
	query := `SELECT id FROM wallet WHERE name = $1 AND url  = $2`

	for _, row := range wallets.Rows[1:] {
		var id string

		err := c.db.Get(&id, query, row.Cells[1].Value, row.Cells[2].Value)
		if err != nil {
			return err
		}

		c.wallets[row.Cells[0].Value] = id
	}

	return nil
}

func (c *accountCommandContext) theFollowingWalletsShouldHave(wallets *gherkin.DataTable) error {
	query := `SELECT id FROM wallet WHERE id = $1`

	for i, cell := range wallets.Rows[0].Cells[1:] {
		query = fmt.Sprintf("%s AND %s=$%d", query, cell.Value, i+2)
	}

	for _, row := range wallets.Rows[1:] {
		var id string
		var args []interface{}

		for _, cell := range row.Cells {
			args = append(args, cell.Value)
		}

		err := c.db.Get(&id, query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *accountCommandContext) theFollowingOperationsShouldBeStored(operations *gherkin.DataTable) error {
	query := `SELECT id FROM operation`
	var where []string

	for i, cell := range operations.Rows[0].Cells[1:] {
		where = append(where, fmt.Sprintf("%s=$%d", cell.Value, i+1))
	}

	if len(where) == 0 {
		return fmt.Errorf("no criteria found")
	}

	query = fmt.Sprintf("%s WHERE %s", query, strings.Join(where, " AND "))

	fmt.Printf("%s", query)

	for _, row := range operations.Rows[1:] {
		var id string
		var args []interface{}

		for k, cell := range row.Cells[1:] {
			if operations.Rows[0].Cells[k+1].Value == "date" {
				args = append(args, parseDateString(cell.Value))

				continue
			}

			args = append(args, cell.Value)
		}

		err := c.db.Get(&id, query, args...)
		if err != nil {
			return err
		}

		c.operations[row.Cells[0].Value] = id
	}

	return nil
}

func (c *accountCommandContext) theFollowingWalletItemsShouldBeStored(operations *gherkin.DataTable) error {
	query := `SELECT id FROM wallet_item`
	var where []string

	for i, cell := range operations.Rows[0].Cells[1:] {
		where = append(where, fmt.Sprintf("%s=$%d", cell.Value, i+1))
	}

	if len(where) == 0 {
		return fmt.Errorf("no criteria found")
	}

	query = fmt.Sprintf("%s WHERE %s", query, strings.Join(where, " AND "))

	fmt.Printf("%s", query)

	for _, row := range operations.Rows[1:] {
		var id string
		var args []interface{}

		for k, cell := range row.Cells[1:] {
			if operations.Rows[0].Cells[k+1].Value == "date" {
				args = append(args, parseDateString(cell.Value))

				continue
			}

			args = append(args, cell.Value)
		}

		err := c.db.Get(&id, query, args...)
		if err != nil {
			return err
		}

		c.walletItems[row.Cells[0].Value] = id
	}

	return nil
}
