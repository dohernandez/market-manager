package bootstrap

import (
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jmoiron/sqlx"
)

type stockCommandContext struct {
	db         *sqlx.DB
	stocksInfo map[string]string
	stocks     map[string]string
}

func RegisterStockCommandContext(s *godog.Suite, db *sqlx.DB) {
	scc := &stockCommandContext{
		db:         db,
		stocksInfo: map[string]string{},
		stocks:     map[string]string{},
	}

	s.Step(`^following stock info should be stored:$`, scc.followingStockInfoShouldBeStored)
	s.Step(`^following stocks should be stored:$`, scc.followingStocksShouldBeStored)
}

func (c *stockCommandContext) followingStockInfoShouldBeStored(stockInfos *gherkin.DataTable) error {
	query := `SELECT id FROM stock_info WHERE name = $1 AND type  = $2`

	for _, row := range stockInfos.Rows[1:] {
		var id string

		err := c.db.Get(&id, query, row.Cells[1].Value, row.Cells[2].Value)
		if err != nil {
			return err
		}

		c.stocksInfo[row.Cells[0].Value] = id
	}

	return nil
}

func (c *stockCommandContext) followingStocksShouldBeStored(stocks *gherkin.DataTable) error {
	query := `SELECT id 
			  FROM stock 
			  WHERE name = $1 
			  AND exchange_id  = $2
			  AND symbol  = $3
			  AND type  = $4
			  AND sector  = $5
			  AND industry  = $6`

	for _, row := range stocks.Rows[1:] {
		var id string

		err := c.db.Get(
			&id,
			query,
			row.Cells[1].Value,
			row.Cells[2].Value,
			row.Cells[3].Value,
			c.stocksInfo[row.Cells[4].Value],
			c.stocksInfo[row.Cells[5].Value],
			c.stocksInfo[row.Cells[6].Value],
		)
		if err != nil {
			return err
		}

		c.stocks[row.Cells[0].Value] = id
	}

	return nil
}
