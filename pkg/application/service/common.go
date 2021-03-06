package service

import (
	"context"

	"golang.org/x/net/html"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// ----------------------------------------------------------------------------------------------------------------------
// stockScrape
// ----------------------------------------------------------------------------------------------------------------------
type stockScrape struct {
	ctx context.Context
	url string
}

// ----------------------------------------------------------------------------------------------------------------------
// UrlBuilder
// ----------------------------------------------------------------------------------------------------------------------
type UrlBuilder interface {
	BuildUrl(stk *stock.Stock) (string, error)
}

// ----------------------------------------------------------------------------------------------------------------------
// HtmlParser
// ----------------------------------------------------------------------------------------------------------------------
type HtmlParser interface {
	Parse(url string) (*html.Node, error)
}
