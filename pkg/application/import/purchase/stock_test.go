package import_purchase

//
//import (
//	"context"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//
//	"errors"
//
//	"github.com/dohernandez/market-manager/pkg/market-manager/market"
//	"github.com/dohernandez/market-manager/pkg/market-manager/market/exchange"
//	"github.com/dohernandez/market-manager/pkg/mocks"
//)
//
//func TestImportStockImport(t *testing.T) {
//	t.Parallel()
//
//	tests := []struct {
//		scenario string
//		function func(*testing.T, *mocks.ReaderMock, *mocks.MarketFinderMock, *mocks.ExchangeFinderMock, *mocks.StockPersisterMock)
//	}{
//		{
//			scenario: "when import stocks success",
//			function: testImportStockSuccess,
//		},
//		{
//			scenario: "when import stocks fail (Market not found)",
//			function: testImportStockFailMarketNotFound,
//		},
//		{
//			scenario: "when import stocks fail (Exchange not found)",
//			function: testImportStockFailExchangeNotFound,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.scenario, func(t *testing.T) {
//			t.Parallel()
//
//			readerMock := mocks.ReaderMock{}
//			marketFinderMock := mocks.MarketFinderMock{}
//			exchangeFinderMock := mocks.ExchangeFinderMock{}
//			stockPersisterMock := mocks.StockPersisterMock{}
//
//			test.function(t, &readerMock, &marketFinderMock, &exchangeFinderMock, &stockPersisterMock)
//		})
//	}
//}
//
//func testImportStockSuccess(t *testing.T, r *mocks.ReaderMock, mf *mocks.MarketFinderMock, ef *mocks.ExchangeFinderMock, sp *mocks.StockPersisterMock) {
//	i := NewImportStock(context.TODO(), r, mf, ef, sp)
//
//	m := market.NewMarket("Market", "symb")
//	e := exchange.NewExchange("Exchange", "symb")
//
//	r.On("Open").Return(nil)
//	r.On("ReadLine").Return([]string{"Stock", "Exchange", "symb"}, nil)
//	r.On("Close").Return(nil)
//
//	mf.On("FindByName", market.Stock).Return(m, nil)
//	mf.On("FindBySymbol", "Exchange").Return(e, nil)
//
//	sp.On("PersistAll", mock.AnythingOfType("[]*stock.Stock")).Return(nil)
//
//	err := i.Import()
//	assert.NoError(t, err)
//
//	r.AssertCalled(t, "Open")
//	r.AssertCalled(t, "ReadLine")
//	mf.AssertCalled(t, "FindByName", market.Stock)
//	ef.AssertCalled(t, "FindBySymbol", "Exchange")
//	sp.AssertCalled(t, "PersistAll", mock.AnythingOfType("[]*stock.Stock"))
//	r.AssertCalled(t, "Close")
//}
//
//func testImportStockFailMarketNotFound(t *testing.T, r *mocks.ReaderMock, mf *mocks.MarketFinderMock, ef *mocks.ExchangeFinderMock, sp *mocks.StockPersisterMock) {
//	i := NewImportStock(context.TODO(), r, mf, ef, sp)
//
//	r.On("Open").Return(nil)
//	r.On("ReadLine").Return([]string{"Stock", "Exchange", "Symb"}, nil)
//	r.On("Close").Return(nil)
//
//	mf.On("FindByName", market.Stock).Return(nil, errors.New("not found"))
//
//	err := i.Import()
//	assert.Error(t, err)
//
//	r.AssertCalled(t, "Open")
//	r.AssertCalled(t, "ReadLine")
//	mf.AssertCalled(t, "FindByName", market.Stock)
//	ef.AssertNotCalled(t, "FindBySymbol")
//	sp.AssertNotCalled(t, "PersistAll")
//	r.AssertCalled(t, "Close")
//}
//
//func testImportStockFailExchangeNotFound(t *testing.T, r *mocks.ReaderMock, mf *mocks.MarketFinderMock, ef *mocks.ExchangeFinderMock, sp *mocks.StockPersisterMock) {
//	i := NewImportStock(context.TODO(), r, mf, ef, sp)
//
//	m := market.NewMarket("Market", "symb")
//
//	r.On("Open").Return(nil)
//	r.On("ReadLine").Return([]string{"Stock", "Exchange", "Symb"}, nil)
//	r.On("Close").Return(nil)
//
//	mf.On("FindByName", market.Stock).Return(m, nil)
//	ef.On("FindBySymbol", "Exchange").Return(nil, errors.New("not found"))
//
//	err := i.Import()
//	assert.Error(t, err)
//
//	r.AssertCalled(t, "Open")
//	r.AssertCalled(t, "ReadLine")
//	mf.AssertCalled(t, "FindByName", market.Stock)
//	ef.AssertCalled(t, "FindBySymbol", "Exchange")
//	sp.AssertNotCalled(t, "PersistAll")
//	r.AssertCalled(t, "Close")
//}
