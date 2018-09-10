package listener

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"

	"github.com/dohernandez/market-manager/pkg/application/util"
)

const linePerFile = 30

func getResourceNumberFromFilePath(fileName string) int {
	reg := regexp.MustCompile(`(^[0-9]{2})+_+(.*)`)
	res := reg.ReplaceAllString(fileName, "${1}")

	n, _ := strconv.Atoi(res)

	return n
}

func getCsvWriterFromResourceImport(resourceStorage util.ResourceStorage, resource, importPath, filesGroup string) (*util.CsvWriter, error) {
	r, err := resourceStorage.FindLastByResourceAndFilesGroup(resource, filesGroup)
	if err != nil {
		return nil, errors.Wrapf(err, "loading latest %q resource import", resource)
	}

	var lines [][]string

	filePath := fmt.Sprintf("%s/%s", importPath, r.FileName)
	rf := util.NewCsvReader(filePath)

	rf.Open()
	defer rf.Close()

	lines, err = rf.ReadAllLines()
	if err != nil {
		return nil, errors.Wrapf(err, "loading previous lines from resource %q", filePath)
	}

	nLines := len(lines)
	if err != nil || nLines >= linePerFile {
		fileNumber := getResourceNumberFromFilePath(r.FileName)

		fName := fmt.Sprintf("%d_%s.csv", fileNumber+1, filesGroup)
		if fileNumber < 10 {
			fName = fmt.Sprintf("0%d_%s.csv", fileNumber+1, filesGroup)
		}

		filePath = fmt.Sprintf("%s/%s", importPath, fName)

		defer func() {
			ir := util.NewResource(resource, fName)

			err := resourceStorage.Persist(ir)
			if err != nil {
				panic(err)
			}
		}()
	}

	return util.NewCsvWriter(filePath), nil
}
