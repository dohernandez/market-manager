package bootstrap

import (
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jmoiron/sqlx"
)

type bankingCommandContext struct {
	db        *sqlx.DB
	transfers map[string]string
}

func RegisterBankingCommandContext(s *godog.Suite, db *sqlx.DB) {
	bcc := &bankingCommandContext{
		db:        db,
		transfers: map[string]string{},
	}

	s.Step(`^following transfers should be stored:$`, bcc.followingTransfersShouldBeStored)
}

func (c *bankingCommandContext) followingTransfersShouldBeStored(transfers *gherkin.DataTable) error {
	query := `SELECT id FROM transfer WHERE from_account = $1 AND to_account  = $2 AND amount  = $3 AND date = $4`

	for _, row := range transfers.Rows[1:] {
		var id string

		a, err := parsePriceString(row.Cells[3].Value)
		if err != nil {
			return err
		}

		err = c.db.Get(&id, query, row.Cells[1].Value, row.Cells[2].Value, a, parseDateString(row.Cells[4].Value))
		if err != nil {
			return err
		}

		c.transfers[row.Cells[0].Value] = id
	}

	return nil
}
