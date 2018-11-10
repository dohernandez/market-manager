package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/dohernandez/market-manager/pkg/application/util"
	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock/dividend"
)

// ----------------------------------------------------------------------------------------------------------------------
// StockDividendMarketChameleon Service
// ----------------------------------------------------------------------------------------------------------------------
type (
	stockDividendMarketChameleon struct {
		ctx        context.Context
		urlBuilder UrlBuilder
		htmlParser HtmlParser
	}

	dividendType string

	stockScrapeMarketChameleonWWWUrlBuilder struct {
		stockScrape
	}

	stockDividendMarketChameleonWWWHtmlParser struct {
		ctx context.Context
	}

	stockScrapeMarketChameleonFileUrlBuilder struct {
		stockScrape
	}

	stockDividendMarketChameleonFileHtmlParser struct {
		ctx context.Context
	}
)

const (
	Future     dividendType = "future_divs"
	Historical              = "historical_divs"
)

// ----------------------------------------------------------------------------------------------------------------------
// UrlBuilder
// ----------------------------------------------------------------------------------------------------------------------
// ----------------------------------------------------------------------------------------------------------------------
// WWW
// ----------------------------------------------------------------------------------------------------------------------
func NewStockScrapeMarketChameleonWWWUrlBuilder(ctx context.Context, url string) UrlBuilder {
	return &stockScrapeMarketChameleonWWWUrlBuilder{
		stockScrape{
			ctx: ctx,
			url: url,
		},
	}
}

func (s *stockScrapeMarketChameleonWWWUrlBuilder) BuildUrl(stk *stock.Stock) (string, error) {
	return fmt.Sprintf("%s/%s/Dividends", s.url, stk.Symbol), nil
}

// ----------------------------------------------------------------------------------------------------------------------
// File
// ----------------------------------------------------------------------------------------------------------------------
func NewStockScrapeMarketChameleonFileUrlBuilder(ctx context.Context, url string) UrlBuilder {
	return &stockScrapeMarketChameleonFileUrlBuilder{
		stockScrape{
			ctx: ctx,
			url: url,
		},
	}
}

func (s *stockScrapeMarketChameleonFileUrlBuilder) BuildUrl(stk *stock.Stock) (string, error) {
	var filePath string

	err := filepath.Walk(s.url, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".html" {
			stockSymbol := util.GeResourceNameFromFilePath(path)
			if stk.Symbol == stockSymbol {
				filePath = path
			}
		}

		return nil
	})
	if err != nil {
		return filePath, err
	}

	return filePath, nil
}

// ----------------------------------------------------------------------------------------------------------------------
// HtmlParser
// ----------------------------------------------------------------------------------------------------------------------
func NewStockDividendMarketChameleonWWWHtmlParser(ctx context.Context) HtmlParser {
	return &stockDividendMarketChameleonWWWHtmlParser{
		ctx: ctx,
	}
}

func (s *stockDividendMarketChameleonWWWHtmlParser) Parse(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, errors.New(
			"marketChameleon said: You do not have permission to view this directory or page using the credentials that you supplied.",
		)
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return root, nil
}

func NewStockDividendMarketChameleonFileHtmlParser(ctx context.Context) HtmlParser {
	return &stockDividendMarketChameleonFileHtmlParser{
		ctx: ctx,
	}
}

func (s *stockDividendMarketChameleonFileHtmlParser) Parse(url string) (*html.Node, error) {
	r, err := os.Open(url)
	if err != nil {
		return nil, err
	}

	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return root, nil
}

func NewStockDividendMarketChameleon(ctx context.Context, urlBuilder UrlBuilder, htmlParser HtmlParser) *stockDividendMarketChameleon {
	return &stockDividendMarketChameleon{
		ctx:        ctx,
		urlBuilder: urlBuilder,
		htmlParser: htmlParser,
	}
}

func (s *stockDividendMarketChameleon) NextFuture(stk *stock.Stock) (dividend.StockDividend, error) {
	url, err := s.urlBuilder.BuildUrl(stk)
	if err != nil {
		return dividend.StockDividend{}, err
	}

	root, err := s.htmlParser.Parse(url)
	if err != nil {
		return dividend.StockDividend{}, err
	}

	sd, err := s.findLastFutureDividend(root)
	if err != nil {
		return dividend.StockDividend{}, err
	}

	logger.FromContext(s.ctx).Debugf("got stock next future dividend %+v from stock %s", sd, stk.Symbol)

	return sd, nil
}

func (s *stockDividendMarketChameleon) findLastFutureDividend(root *html.Node) (dividend.StockDividend, error) {
	rows, err := s.findDividendsRows(root, Future)
	if err != nil {
		return dividend.StockDividend{}, err
	}

	var stkDividend dividend.StockDividend

	lrow := rows[len(rows)-1]

	tds := scrape.FindAll(lrow, scrape.ByTag(atom.Td))

	s.marshalStockFutureDividend(tds, &stkDividend)

	return stkDividend, nil
}

func (s *stockDividendMarketChameleon) findDividendsRows(root *html.Node, dt dividendType) ([]*html.Node, error) {
	if dt != Future && dt != Historical {
		return nil, errors.Errorf("There aren't any dividend table with this ID [%s]", dt)
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Table && scrape.Attr(n, "id") == fmt.Sprint(dt) {
			return true
		}

		return false
	}

	tableDividendMarketChameleon, ok := scrape.Find(root, matcher)
	if !ok {
		return nil, errors.Errorf("%s not found", dt)
	}

	// define a matcher for tbody > tr
	matcher = func(n *html.Node) bool {
		if n.DataAtom == atom.Tr && n.Parent.DataAtom == atom.Tbody {
			return true
		}

		return false
	}

	rows := scrape.FindAll(tableDividendMarketChameleon, matcher)
	if len(rows) == 0 {
		return nil, errors.New("There aren't any future dividend")
	}

	return rows, nil
}

func (s *stockDividendMarketChameleon) marshalStockFutureDividend(tds []*html.Node, stkDividend *dividend.StockDividend) {
	for i, td := range tds {
		switch i {
		case 0:
			exDates := strings.Split(scrape.Text(td), " - ")
			if len(exDates) == 1 {
				stkDividend.ExDate = s.parseDateString(exDates[0])
			} else {
				stkDividend.ExDate = s.parseDateString(exDates[1])

			}
		case 1:
			rDate := scrape.Text(td)
			if rDate != "" {
				stkDividend.RecordDate = s.parseDateString(rDate)
			}
		case 2:
			pDate := scrape.Text(td)
			if pDate != "" {
				stkDividend.PaymentDate = s.parseDateString(pDate)
			}
		case 3:
			status := scrape.Text(td)
			switch status {
			case "Projected":
				stkDividend.Status = dividend.Projected
			case "Announced":
				stkDividend.Status = dividend.Announced
			default:
				panic(fmt.Sprintf("Status not supported [%s]", status))
			}
		case 5:
			stkDividend.Amount = mm.ValueDollarFromString(scrape.Text(td))
		}
	}
}

// parseDateString - parse a potentially partial date string to Time
func (s *stockDividendMarketChameleon) parseDateString(dt string) time.Time {
	if dt == "" {
		return time.Now()
	}

	t, _ := time.Parse("2-Jan-2006", dt)

	return t
}

func (s *stockDividendMarketChameleon) Future(stk *stock.Stock) ([]dividend.StockDividend, error) {
	url, err := s.urlBuilder.BuildUrl(stk)
	if err != nil {
		return nil, err
	}

	root, err := s.htmlParser.Parse(url)
	if err != nil {
		return nil, err
	}

	sd, err := s.findAllFutureDividend(root)
	if err != nil {
		return nil, err
	}

	logger.FromContext(s.ctx).Debugf("got stock future dividend %+v from stock %s", sd, stk.Symbol)

	return sd, nil
}

func (s *stockDividendMarketChameleon) findAllFutureDividend(root *html.Node) ([]dividend.StockDividend, error) {
	rows, err := s.findDividendsRows(root, Future)
	if err != nil {
		return nil, err
	}

	var stkDividends []dividend.StockDividend

	for _, row := range rows {
		var stkDividend dividend.StockDividend
		tds := scrape.FindAll(row, scrape.ByTag(atom.Td))

		s.marshalStockFutureDividend(tds, &stkDividend)

		stkDividends = append(stkDividends, stkDividend)
	}

	return stkDividends, nil
}

func (s *stockDividendMarketChameleon) Historical(stk *stock.Stock, fromDate time.Time) ([]dividend.StockDividend, error) {
	url, err := s.urlBuilder.BuildUrl(stk)
	if err != nil {
		return nil, err
	}

	root, err := s.htmlParser.Parse(url)
	if err != nil {
		return nil, err
	}

	sd, err := s.findAllHistoricalDividend(root, fromDate)
	if err != nil {
		return nil, err
	}

	logger.FromContext(s.ctx).Debugf("got stock historical dividend %+v from stock %s", sd, stk.Symbol)

	return sd, nil
}

func (s *stockDividendMarketChameleon) findAllHistoricalDividend(root *html.Node, fromDate time.Time) ([]dividend.StockDividend, error) {
	rows, err := s.findDividendsRows(root, Historical)
	if err != nil {
		return nil, err
	}

	var stkDividends []dividend.StockDividend

	for _, row := range rows {
		stkDividend := dividend.StockDividend{
			Status: dividend.Payed,
		}
		tds := scrape.FindAll(row, scrape.ByTag(atom.Td))

		s.marshalStockFutureHistorical(tds, &stkDividend)

		if stkDividend.ExDate.After(fromDate) {
			stkDividends = append(stkDividends, stkDividend)
		}
	}

	return stkDividends, nil
}

func (s *stockDividendMarketChameleon) marshalStockFutureHistorical(tds []*html.Node, stkDividend *dividend.StockDividend) {
	now := time.Now()

	for i, td := range tds {
		switch i {
		case 0:
			exDate := scrape.Text(td)
			if exDate != "" {
				stkDividend.ExDate = s.parseDateString(exDate)
			}
		case 1:
			rDate := scrape.Text(td)
			if rDate != "" {
				stkDividend.RecordDate = s.parseDateString(rDate)
			}
		case 2:
			pDate := scrape.Text(td)
			if pDate != "" {
				stkDividend.PaymentDate = s.parseDateString(pDate)

				if stkDividend.PaymentDate.After(now) {
					stkDividend.Status = dividend.Announced
				}
			}
		case 4:
			stkDividend.Amount = mm.ValueDollarFromString(scrape.Text(td))
		case 6:
			stkDividend.ChangeFromPrev = s.sanitizePercentage(scrape.Text(td))
		case 7:
			stkDividend.ChangeFromPrevYear = s.sanitizePercentage(scrape.Text(td))
		case 8:
			stkDividend.Prior12MonthsYield = s.sanitizePercentage(scrape.Text(td))
		}
	}
}

func (s *stockDividendMarketChameleon) sanitizePercentage(percentage string) float64 {
	p := strings.Replace(percentage, "%", "", 1)
	p = strings.Replace(p, "+", "", 1)

	pf, _ := strconv.ParseFloat(p, 64)

	return pf
}
