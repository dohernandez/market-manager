package mocks

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/exchange"
	"github.com/stretchr/testify/mock"
)

type ExchangeFinderMock struct {
	mock.Mock
}

func (m *ExchangeFinderMock) FindBySymbol(n string) (*exchange.Exchange, error) {
	args := m.Called(n)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*exchange.Exchange), args.Error(1)
}
