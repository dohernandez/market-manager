package mocks

import (
	"github.com/dohernandez/market-manager/pkg/market-manager/market"
	"github.com/stretchr/testify/mock"
)

type MarketFinderMock struct {
	mock.Mock
}

func (m *MarketFinderMock) FindByName(n string) (*market.Market, error) {
	args := m.Called(n)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*market.Market), args.Error(1)
}
