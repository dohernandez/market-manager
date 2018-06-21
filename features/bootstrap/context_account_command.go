package bootstrap

import (
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
