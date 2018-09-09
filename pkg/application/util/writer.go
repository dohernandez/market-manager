package util

import (
	"bufio"
	"encoding/csv"
	"os"
)

type Writer interface {
	Open() error
	Close() error
	WriteAllLines(lines [][]string) error
	Flush()
}

type CsvWriter struct {
	csvFileName string

	file   *os.File
	writer *csv.Writer
}

func NewCsvWriter(csvFileName string) *CsvWriter {
	return &CsvWriter{
		csvFileName: csvFileName,
	}
}

func (r *CsvWriter) Open() error {
	if r.file != nil {
		return ErrFileOpen
	}

	csvFile, err := os.OpenFile(r.csvFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	r.file = csvFile
	r.writer = csv.NewWriter(bufio.NewWriter(csvFile))

	return nil
}

func (r *CsvWriter) Close() error {
	err := r.file.Close()

	r.file = nil
	r.writer = nil

	return err
}

func (r *CsvWriter) WriteAllLines(lines [][]string) error {
	return r.writer.WriteAll(lines)
}

func (r *CsvWriter) Flush() {
	r.writer.Flush()
}
