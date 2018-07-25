package service

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/dohernandez/go-quote"
	gf "github.com/dohernandez/googlefinance-client-go"
	"github.com/dohernandez/market-manager/pkg/infrastructure/client/go-iex"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// ----------------------------------------------------------------------------------------------------------------------
// BasicStockPrice Service
// ----------------------------------------------------------------------------------------------------------------------
type (
	basicStockPrice struct {
		ctx       context.Context
		iexClient *iex.Client
	}
)

func NewBasicStockPrice(ctx context.Context, iexClient *iex.Client) *basicStockPrice {
	return &basicStockPrice{
		ctx:       ctx,
		iexClient: iexClient,
	}
}

func (s *basicStockPrice) Price(stk *stock.Stock) (stock.Price, error) {
	method := "closedPriceFromYahoo"

	p, err := s.closedPriceFromYahoo(stk)
	if err != nil {
		logger.FromContext(s.ctx).WithError(err).Debugf("failed %s for stock %q", method, stk.Symbol)
		time.Sleep(5 * time.Second)
		method = "closedPriceFromGoogle"

		p, err = s.closedPriceFromGoogle(stk)
		if err != nil {
			logger.FromContext(s.ctx).WithError(err).Debugf("failed %s for stock %q", method, stk.Symbol)
			time.Sleep(5 * time.Second)
			method = "closedPriceFromIEXTrading"

			p, err = s.closedPriceFromIEXTrading(stk)
			if err != nil {
				if err == mm.ErrNotFound {
					return stock.Price{}, err
				}

				return stock.Price{}, errors.WithStack(err)
			}
		}
	}
	logger.FromContext(s.ctx).Debugf("got price %+v from stock %s with method %s", p, stk.Symbol, method)

	return p, nil
}

func (s *basicStockPrice) closedPriceFromYahoo(stk *stock.Stock) (stock.Price, error) {
	endDate := time.Now()
	startDate := endDate.Add(-24 * time.Hour)

	q, err := quote.NewQuoteFromYahoo(stk.Symbol, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), quote.Daily, true)
	if err != nil {
		return stock.Price{}, err
	}

	return stock.Price{
		Date:   q.Date[0],
		Close:  q.Close[0],
		High:   q.High[0],
		Low:    q.Low[0],
		Open:   q.Open[0],
		Volume: int64(q.Volume[0]),
		Change: q.Close[0] - q.Open[0],
	}, nil
}

func (s *basicStockPrice) closedPriceFromGoogle(stk *stock.Stock) (stock.Price, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gps, err := gf.GetPrices(ctx, &gf.Query{
		P: "2d",
		I: "86400",
		X: stk.Exchange.Symbol,
		Q: stk.Symbol,
	})
	if err != nil {
		return stock.Price{}, err
	}

	if len(gps) == 0 {
		return stock.Price{}, errors.Errorf("symbol '%s' not found\n", stk.Symbol)
	}

	p := gps[len(gps)-1]

	return stock.Price{
		Date:   p.Date,
		Close:  p.Close,
		High:   p.High,
		Low:    p.Low,
		Open:   p.Open,
		Volume: p.Volume,
		Change: p.Close - p.Open,
	}, nil
}

func (s *basicStockPrice) closedPriceFromIEXTrading(stk *stock.Stock) (stock.Price, error) {
	q, err := s.iexClient.Quote.Get(stk.Symbol)
	if err != nil {
		return stock.Price{}, err
	}
	return stock.Price{
		//Date:   q.LatestUpdate,
		Close:  q.Close,
		High:   q.High,
		Low:    q.Low,
		Open:   q.Open,
		Volume: q.Volume,
		Change: q.Close - q.Open,
	}, nil
}

// ----------------------------------------------------------------------------------------------------------------------
// YahooScraperStockPrice Service
// ----------------------------------------------------------------------------------------------------------------------
type (
	yahooScraperStockPrice struct {
		stockScrape
	}
)

func NewYahooScrapeStockPrice(ctx context.Context, url string) *yahooScraperStockPrice {
	return &yahooScraperStockPrice{
		stockScrape: stockScrape{
			ctx: ctx,
			url: url,
		},
	}
}

func (s *yahooScraperStockPrice) Price(stk *stock.Stock) (stock.Price, error) {
	url := fmt.Sprintf("%s/%s?p=%s", s.url, stk.Symbol, stk.Symbol)

	resp, err := http.Get(url)
	if err != nil {
		return stock.Price{}, err
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return stock.Price{}, err
	}

	p := stock.Price{
		Date: time.Now(),
	}

	err = s.marshalQuoteHeaderInfo(root, &p)
	if err != nil {
		return stock.Price{}, err
	}

	err = s.marshalQuoteSummary(root, &p)
	if err != nil {
		return stock.Price{}, err
	}

	logger.FromContext(s.ctx).Debugf("got price %+v from stock %s", p, stk.Symbol)

	return p, nil
}

func (s *yahooScraperStockPrice) marshalQuoteHeaderInfo(root *html.Node, p *stock.Price) error {
	// define a matcher
	matcherQuoteHeaderInfo := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && scrape.Attr(n, "id") == "quote-header-info" {
			return true
		}

		return false
	}
	quoteHeaderInfoDiv, ok := scrape.Find(root, matcherQuoteHeaderInfo)
	if !ok {
		return errors.New("Marshal quote header info. Quote header not found")
	}

	err := s.marshalQuoteHeaderInfoClosePrice(quoteHeaderInfoDiv, p)
	if err != nil {
		return err
	}

	err = s.marshalQuoteHeaderInfoChangePrice(quoteHeaderInfoDiv, p)
	if err != nil {
		return err
	}

	return nil
}

func (s *yahooScraperStockPrice) marshalQuoteHeaderInfoClosePrice(root *html.Node, p *stock.Price) error {
	// define a matcher
	matcherQuoteClosedPrice := func(n *html.Node) bool {
		if n.DataAtom == atom.Span && scrape.Attr(n, "data-reactid") == "21" {
			return true
		}

		return false
	}

	quoteClosedPriceSpan, ok := scrape.Find(root, matcherQuoteClosedPrice)
	if !ok {
		return errors.New("Marshal quote header info. Closed price not found")
	}

	p.Close, _ = strconv.ParseFloat(scrape.Text(quoteClosedPriceSpan), 64)

	return nil
}

func (s *yahooScraperStockPrice) marshalQuoteHeaderInfoChangePrice(root *html.Node, p *stock.Price) error {
	// define a matcher
	matcherQuoteChangePrice := func(n *html.Node) bool {
		if n.DataAtom == atom.Span && scrape.Attr(n, "data-reactid") == "23" {
			return true
		}

		return false
	}

	quoteChangePriceSpan, ok := scrape.Find(root, matcherQuoteChangePrice)
	if !ok {
		return errors.New("Marshal quote header info. Closed price not found")
	}

	re := regexp.MustCompile(`^(.*) .*`)
	matches := re.FindStringSubmatch(scrape.Text(quoteChangePriceSpan))

	if len(matches) > 0 {
		p.Change, _ = strconv.ParseFloat(matches[1], 64)
	}

	return nil
}

func (s *yahooScraperStockPrice) marshalQuoteSummary(root *html.Node, p *stock.Price) error {
	// define a matcher for summary div
	matcherQuoteSummary := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && scrape.Attr(n, "id") == "quote-summary" {
			return true
		}

		return false
	}
	quoteSummaryDiv, ok := scrape.Find(root, matcherQuoteSummary)
	if !ok {
		return errors.New("Price not found. Quote summary not found")
	}

	err := s.marshalQuoteSummaryLeftTable(quoteSummaryDiv, p)
	if err != nil {
		return err
	}

	err = s.marshalQuoteSummaryRightTable(quoteSummaryDiv, p)
	if err != nil {
		return err
	}

	return nil
}

func (s *yahooScraperStockPrice) marshalQuoteSummaryLeftTable(root *html.Node, p *stock.Price) error {
	// define a matcher left table
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.Div && scrape.Attr(n, "data-test") == "left-summary-table" {
			return true
		}

		return false
	}

	leftSummaryTable, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("Price not found. Quote summary left table not found")
	}

	rows := scrape.FindAll(leftSummaryTable, scrape.ByTag(atom.Tr))

	for i, row := range rows {
		switch i {
		case 1:
			p.Open, _ = strconv.ParseFloat(scrape.Text(row.FirstChild.NextSibling), 64)
		case 4:
			highLowDay := strings.Split(scrape.Text(row.FirstChild.NextSibling), " - ")

			if len(highLowDay) == 2 {
				p.Low, _ = strconv.ParseFloat(highLowDay[0], 64)
				p.High, _ = strconv.ParseFloat(highLowDay[1], 64)
			}
		case 5:
			highLow52week := strings.Split(scrape.Text(row.FirstChild.NextSibling), " - ")

			if len(highLow52week) == 2 {
				p.Low52Week, _ = strconv.ParseFloat(highLow52week[0], 64)
				p.High52Week, _ = strconv.ParseFloat(highLow52week[1], 64)
			}
		case 6:
			vStr := s.sanitizeVolume(scrape.Text(row.FirstChild.NextSibling))
			p.Volume, _ = strconv.ParseInt(vStr, 10, 64)
		}
	}

	return nil
}

func (s *yahooScraperStockPrice) sanitizeVolume(v string) string {
	return strings.Replace(v, ",", "", 1)
}

func (s *yahooScraperStockPrice) marshalQuoteSummaryRightTable(root *html.Node, p *stock.Price) error {
	// define a matcher left table
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.Div && scrape.Attr(n, "data-test") == "right-summary-table" {
			return true
		}

		return false
	}

	rightSummaryTable, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("Price not found. Quote summary right table not found")
	}

	rows := scrape.FindAll(rightSummaryTable, scrape.ByTag(atom.Tr))

	for i, row := range rows {
		switch i {
		case 2:
			p.PER, _ = strconv.ParseFloat(scrape.Text(row.FirstChild.NextSibling), 64)
		case 3:
			p.EPS, _ = strconv.ParseFloat(scrape.Text(row.FirstChild.NextSibling), 64)
		}
	}

	return nil
}
