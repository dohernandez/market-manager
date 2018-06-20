package bootstrap

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jmoiron/sqlx"
)

type stockCommandContext struct {
	db         *sqlx.DB
	stocksPath string
	stocksInfo map[string]string
	stockIs    map[string]string
}

func StockCommandContext(s *godog.Suite, db *sqlx.DB, stocksPath string) {
	scc := &stockCommandContext{
		db:         db,
		stocksPath: stocksPath,
		stocksInfo: map[string]string{},
		stockIs:    map[string]string{},
	}

	s.BeforeScenario(func(i interface{}) {
		directory := scc.stocksPath

		dirClean, _ := os.Open(directory)
		dirFiles, _ := dirClean.Readdir(0)

		// Loop over the directory's files.
		for index := range dirFiles {
			fileHere := dirFiles[index]

			// Get name of file and its full path.
			nameHere := fileHere.Name()
			if nameHere != ".gitkeep" {
				fullPath := directory + nameHere

				// Remove the file.
				os.Remove(fullPath)
			}
		}
	})

	s.Step(`^I add a new csv file "([^"]*)" to the stock import folder with the following lines$`, scc.iAddANewCsvFileToTheStockImportFolderWithTheFollowingLines)
	s.Step(`^I run a command "([^"]*)" with args "([^"]*)"$`, scc.iRunACommand)
	s.Step(`^following stock info should be stored:$`, scc.followingStockInfoShouldBeStored)
	s.Step(`^following stocks should be stored:$`, scc.followingStocksShouldBeStored)
}

func (c *stockCommandContext) iAddANewCsvFileToTheStockImportFolderWithTheFollowingLines(filename string, lines *gherkin.DataTable) error {
	file, err := os.Create(fmt.Sprintf("%s/%s", c.stocksPath, filename))
	if err != nil {
		return fmt.Errorf("cannot create file %s", filename)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range lines.Rows[1:] {
		err := writer.Write([]string{
			row.Cells[0].Value,
			row.Cells[1].Value,
			row.Cells[2].Value,
			row.Cells[3].Value,
			row.Cells[4].Value,
			row.Cells[5].Value,
		})
		if err != nil {
			return fmt.Errorf("cannot write to file %s", filename)
		}
	}

	return nil
}

func (c *stockCommandContext) iRunACommand(command, args string) error {
	cArgs := strings.Split(args, " ")
	cmd := exec.Command(command, cArgs...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("running command %s. Error: %s", command, err.Error())
	}

	return nil
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

		c.stockIs[row.Cells[0].Value] = id
	}

	return nil
}
