package service

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"fmt"
	"net/http"

	"time"

	"github.com/dohernandez/market-manager/pkg/infrastructure/logger"
	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// ----------------------------------------------------------------------------------------------------------------------
// StockPriceVolatility Service
// ----------------------------------------------------------------------------------------------------------------------
type (
	marketChameleonStockPriceVolatility struct {
		stockScrape
	}
)

func NewMarketChameleonStockPriceVolatility(ctx context.Context, url string) *marketChameleonStockPriceVolatility {
	return &marketChameleonStockPriceVolatility{
		stockScrape: stockScrape{
			ctx: ctx,
			url: url,
		},
	}
}

func (s *marketChameleonStockPriceVolatility) PriceVolatility(stk *stock.Stock) (stock.PriceVolatility, error) {
	url := fmt.Sprintf("%s/%s?p=%s", s.url, stk.Symbol, stk.Symbol)

	resp, err := http.Get(url)
	if err != nil {
		return stock.PriceVolatility{}, err
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return stock.PriceVolatility{}, err
	}

	stkPriceVolatility := stock.PriceVolatility{
		Date: time.Now(),
	}

	err = s.marshalVolatility(root, &stkPriceVolatility)
	if err != nil {
		return stock.PriceVolatility{}, err
	}

	logger.FromContext(s.ctx).Debugf("got price volatility %+v from stock %s", stkPriceVolatility, stk.Symbol)

	return stkPriceVolatility, nil
}

func (s *marketChameleonStockPriceVolatility) marshalVolatility(root *html.Node, stkPriceVolatility *stock.PriceVolatility) error {
	// define a matcher
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && scrape.Attr(n, "class") == "symov_stat_box _c" {
			if n.FirstChild != nil && n.FirstChild.NextSibling != nil && scrape.Text(n.FirstChild.NextSibling) == "Volatility" {
				return true
			}
		}

		return false
	}

	divMarketChameleonVolatilityDiv, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("Stock volatility not found")
	}

	divs := scrape.FindAll(divMarketChameleonVolatilityDiv, scrape.ByClass("flex_container_between"))

	for i, div := range divs {
		d, ok := scrape.Find(div, scrape.ByClass("datatag"))
		if !ok {
			continue
		}

		switch i {
		case 1:
			stkPriceVolatility.HV20Day, _ = strconv.ParseFloat(scrape.Text(d), 64)
		case 2:
			stkPriceVolatility.HV52Week, _ = strconv.ParseFloat(scrape.Text(d), 64)
		}
	}

	return nil
}
