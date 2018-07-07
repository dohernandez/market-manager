package mocks

import (
	"io"

	"github.com/stretchr/testify/mock"
)

type ReaderMock struct {
	mock.Mock

	eof int
}

func (m *ReaderMock) Open() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ReaderMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ReaderMock) ReadLine() (record []string, err error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	if m.eof == 1 {
		return nil, io.EOF
	}

	m.eof++

	return args.Get(0).([]string), args.Error(1)
}
