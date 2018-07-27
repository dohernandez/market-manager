package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// ----------------------------------------------------------------------------------------------------------------------
// StockSummaryMarketChameleon Service
// ----------------------------------------------------------------------------------------------------------------------
type (
	stockSummaryMarketChameleon struct {
		stockScrape
	}
)

func NewStockSummaryMarketChameleon(ctx context.Context, url string) *stockSummaryMarketChameleon {
	return &stockSummaryMarketChameleon{
		stockScrape: stockScrape{
			ctx: ctx,
			url: url,
		},
	}
}

func (s *stockSummaryMarketChameleon) Summary(stk *stock.Stock) (stock.Summary, error) {
	url := fmt.Sprintf("%s/%s", s.url, stk.Symbol)

	resp, err := http.Get(url)
	if err != nil {
		return stock.Summary{}, err
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return stock.Summary{}, err
	}

	var stkSummary stock.Summary

	err = s.marshalStockInfo(root, &stkSummary)
	if err != nil {
		return stock.Summary{}, err
	}

	logger.FromContext(s.ctx).Debugf("got summary %+v from stock %s", stkSummary, stk.Symbol)

	return stkSummary, nil
}

func (s *stockSummaryMarketChameleon) marshalStockInfo(root *html.Node, stkSummary *stock.Summary) error {
	// define a matcher
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && scrape.Attr(n, "class") == "symov_stat_box symov_info_box _c" {
			return true
		}

		return false
	}

	divMarketChameleonInfoDiv, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("Stock info not found")
	}

	divs := scrape.FindAll(divMarketChameleonInfoDiv, scrape.ByClass("flex_container_between"))

	for i, div := range divs {
		d, ok := scrape.Find(div, scrape.ByClass("datatag"))
		if !ok {
			continue
		}

		switch i {
		case 0:
			stkSummary.Type = scrape.Text(d)
		case 1:
			stkSummary.Sector = scrape.Text(d)
		case 2:
			stkSummary.Industry = scrape.Text(d)
		}
	}

	return nil
}
