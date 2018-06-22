package bootstrap

import (
	"fmt"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jmoiron/sqlx"
)

type accountCommandContext struct {
	db      *sqlx.DB
	wallets map[string]string
}

func RegisterAccountCommandContext(s *godog.Suite, db *sqlx.DB) {
	acc := &accountCommandContext{
		db:      db,
		wallets: map[string]string{},
	}

	s.Step(`^following wallets should be stored:$`, acc.followingWalletsShouldBeStored)
	s.Step(`^the wallets should have:$`, acc.theWalletsShouldHave)
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

func (c *accountCommandContext) theWalletsShouldHave(wallets *gherkin.DataTable) error {
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
