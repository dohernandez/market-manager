package bootstrap

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
)

type csvFileContext struct {
	stocksPath    string
	walletPath    string
	transfersPath string
}

func RegisterCsvFileContext(s *godog.Suite, stocksPath, walletPath, transfersPath string) {
	fc := &csvFileContext{
		stocksPath:    stocksPath,
		walletPath:    walletPath,
		transfersPath: transfersPath,
	}

	s.BeforeScenario(func(i interface{}) {
		fc.cleanDir(fc.stocksPath)
		fc.cleanDir(fc.walletPath)
		fc.cleanDir(fc.transfersPath)
	})

	s.Step(`^I add a new csv file "([^"]*)" to the "([^"]*)" import folder with the following lines$`, fc.iAddANewCsvFileToTheStockImportFolderWithTheFollowingLines)
}

func (c *csvFileContext) cleanDir(directory string) {
	dirClean, _ := os.Open(directory)
	dirFiles, _ := dirClean.Readdir(0)

	// Loop over the directory's files.
	for index := range dirFiles {
		fileHere := dirFiles[index]

		// Get name of file and its full path.
		nameHere := fileHere.Name()
		if nameHere != ".gitkeep" {
			fullPath := fmt.Sprintf("%s/%s", directory, nameHere)

			// Remove the file.
			err := os.Remove(fullPath)
			if err != nil {
				panic(fmt.Errorf("can not remove file %s", fullPath))
			}
		}
	}
}

func (c *csvFileContext) iAddANewCsvFileToTheStockImportFolderWithTheFollowingLines(filename, folder string, lines *gherkin.DataTable) error {
	var basePath string
	switch folder {
	case "stock":
		basePath = c.stocksPath
	case "wallet":
		basePath = c.walletPath
	case "transfer":
		basePath = c.transfersPath
	default:
		return fmt.Errorf("folder not allowed")
	}

	file, err := os.Create(fmt.Sprintf("%s/%s", basePath, filename))
	if err != nil {
		return fmt.Errorf("cannot create file %s", filename)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range lines.Rows[1:] {
		var line []string
		for _, cell := range row.Cells {
			line = append(line, cell.Value)
		}

		err := writer.Write(line)
		if err != nil {
			return fmt.Errorf("cannot write to file %s", filename)
		}
	}

	return nil
}
