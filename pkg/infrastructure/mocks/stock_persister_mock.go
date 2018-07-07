package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

type StockPersisterMock struct {
	mock.Mock
}

func (m *StockPersisterMock) PersistAll(ss []*stock.Stock) error {
	args := m.Called(ss)

	return args.Error(0)
}
