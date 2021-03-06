package util

import (
	"bufio"
	"encoding/csv"
	"os"
)

type Reader interface {
	Open() error
	Close() error
	ReadLine() (record []string, err error)
	ReadAllLines() (record [][]string, err error)
}

type CsvReader struct {
	csvFileName string

	file   *os.File
	reader *csv.Reader
}

func NewCsvReader(csvFileName string) *CsvReader {
	return &CsvReader{
		csvFileName: csvFileName,
	}
}

func (r *CsvReader) Open() error {
	csvFile, err := os.Open(r.csvFileName)
	if err != nil {
		return err
	}

	r.file = csvFile
	r.reader = csv.NewReader(bufio.NewReader(csvFile))

	return nil
}

func (r *CsvReader) Close() error {
	err := r.file.Close()

	r.file = nil
	r.reader = nil

	return err
}

func (r *CsvReader) ReadLine() (record []string, err error) {
	return r.reader.Read()
}

func (r *CsvReader) ReadAllLines() (record [][]string, err error) {
	return r.reader.ReadAll()
}
