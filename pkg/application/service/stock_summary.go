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

	if resp.StatusCode == http.StatusForbidden {
		return stock.Summary{}, errors.New(
			"marketChameleon said: You do not have permission to view this directory or page using the credentials that you supplied.",
		)
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

	logger.FromContext(s.ctx).Debugf("stockSummaryMarketChameleon got summary %+v from stock %s", stkSummary, stk.Symbol)

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

// ----------------------------------------------------------------------------------------------------------------------
// StockSummaryYahoo Service
// ----------------------------------------------------------------------------------------------------------------------
type (
	stockSummaryYahoo struct {
		stockScrape
	}
)

func NewStockSummaryYahoo(ctx context.Context, url string) *stockSummaryYahoo {
	return &stockSummaryYahoo{
		stockScrape: stockScrape{
			ctx: ctx,
			url: url,
		},
	}
}

func (s *stockSummaryYahoo) Summary(stk *stock.Stock) (stock.Summary, error) {
	stkSymbol := stk.Symbol
	if stk.Exchange.Symbol == "TSX" {
		stkSymbol = stk.Symbol + ".TO"
	}

	url := fmt.Sprintf("%s/%s/profile?p=%s", s.url, stkSymbol, stk.Symbol)

	resp, err := http.Get(url)
	if err != nil {
		return stock.Summary{}, err
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return stock.Summary{}, err
	}

	var stkSummary stock.Summary

	err = s.marshalStockName(root, &stkSummary)
	if err != nil {
		return stock.Summary{}, err
	}

	err = s.marshalStockInfo(root, &stkSummary)
	if err != nil {
		return stock.Summary{}, err
	}

	err = s.marshalStockDescription(root, &stkSummary)
	if err != nil {
		return stock.Summary{}, err
	}

	logger.FromContext(s.ctx).Debugf("stockSummaryYahoo got summary %+v from stock %s", stkSummary, stk.Symbol)

	return stkSummary, nil
}

func (s *stockSummaryYahoo) marshalStockName(root *html.Node, stkSummary *stock.Summary) error {
	// define a matcher
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.H3 && scrape.Attr(n, "class") == "Fz(m) Mb(10px)" {
			if n.Parent.DataAtom == atom.Div && scrape.Attr(n.Parent, "class") == "qsp-2col-profile Mt(10px) smartphone_Mt(20px) Lh(1.7)" {
				if n.Parent.Parent.DataAtom == atom.Div && scrape.Attr(n.Parent.Parent, "class") == "asset-profile-container" {
					if n.Parent.Parent.Parent.DataAtom == atom.Section && scrape.Attr(n.Parent.Parent.Parent, "class") == "Pb(30px) smartphone_Px(20px)" {
						return true
					}
				}
			}
		}

		return false
	}

	stockNameH3, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("Stock name info not found")
	}

	stkSummary.Name = scrape.Text(stockNameH3)

	return nil
}

func (s *stockSummaryYahoo) marshalStockInfo(root *html.Node, stkSummary *stock.Summary) error {
	// define a matcher
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.P && scrape.Attr(n, "class") == "D(ib) Va(t)" {
			if n.Parent.DataAtom == atom.Div && scrape.Attr(n.Parent, "class") == "Mb(25px)" {
				if n.Parent.Parent.DataAtom == atom.Div && scrape.Attr(n.Parent.Parent, "class") == "qsp-2col-profile Mt(10px) smartphone_Mt(20px) Lh(1.7)" {
					if n.Parent.Parent.Parent.DataAtom == atom.Div && scrape.Attr(n.Parent.Parent.Parent, "class") == "asset-profile-container" {
						return true
					}
				}
			}
		}

		return false
	}

	infoP, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("Stock info not found")
	}

	for c := infoP.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == atom.Span && scrape.Text(c) == "Sector" {
			var found bool

			for c = c.NextSibling; c != nil; c = c.NextSibling {
				if c.DataAtom == atom.Strong {
					stkSummary.Sector = scrape.Text(c)

					found = true
					break
				}
			}

			if !found {
				return errors.New("Stock sector info not found")
			}
		}

		if c.DataAtom == atom.Span && scrape.Text(c) == "Industry" {
			var found bool

			for c = c.NextSibling; c != nil; c = c.NextSibling {
				if c.DataAtom == atom.Strong {
					stkSummary.Industry = scrape.Text(c)

					found = true
					break
				}
			}

			if !found {
				return errors.New("Stock sector info not found")
			}
		}
	}

	return nil
}

func (s *stockSummaryYahoo) marshalStockDescription(root *html.Node, stkSummary *stock.Summary) error {
	// define a matcher
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Section && scrape.Attr(n, "class") == "quote-sub-section Mt(30px)" {
			if n.Parent.DataAtom == atom.Section && scrape.Attr(n.Parent, "class") == "Pb(30px) smartphone_Px(20px)" {
				return true
			}
		}

		return false
	}

	descriptionSection, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("Stock description not found")
	}

	c := descriptionSection.FirstChild

	if scrape.Text(c) != "Description" {
		return errors.New("Stock description info not found")
	}

	c = c.NextSibling

	stkSummary.Description = scrape.Text(c)

	return nil
}
