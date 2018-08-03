package cmd

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/urfave/cli"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// Scraper ...
type (
	Scraper struct {
		*Base
	}

	marketChameleon struct {
		Stock         stock.Stock
		StockDividend struct {
			NextExDate     string
			Amount         string
			ChangePrevYear string
			ForwardYield   string
		}
		StockInfo struct {
			Type     string
			Sector   string
			Industry string
		}
		StockVolatility struct {
			HV20Day  string
			HV52Week string
		}
	}
)

// NewScraper constructs ApiCMD
func NewScraper(baseCMD *Base) *Scraper {
	return &Scraper{
		Base: baseCMD,
	}
}

// DividendGrowthStocks from dividend-calculator.com
func (cmd *Scraper) DividendGrowthStocks(cliCtx *cli.Context) error {
	var stks []stock.Stock

	// Getting stocks from http://www.dividend-calculator.com/dividend-growth-stocks-list.php
	resp, err := http.Get("http://www.dividend-calculator.com/dividend-growth-stocks-list.php")
	if err != nil {
		panic(err)
	}

	//buf := new(bytes.Buffer)
	//buf.ReadFrom(resp.Body)
	//newStr := buf.String()
	//
	//fmt.Printf(newStr)

	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	// define a matcher
	matcher := func(n *html.Node) bool {
		// must check for nil values
		if n.DataAtom == atom.Tbody {
			if n.Parent.DataAtom == atom.Table && scrape.Attr(n.Parent, "id") == "sortedtable" {
				return true
			}
		}

		return false
	}

	tBodyTableDividendGrowthStocksList, ok := scrape.Find(root, matcher)
	if !ok {
		return errors.New("dividend growth stocks list not found")
	}

	trs := scrape.FindAll(tBodyTableDividendGrowthStocksList, scrape.ByTag(atom.Tr))

	for _, tr := range trs {
		tds := scrape.FindAll(tr, scrape.ByTag(atom.Td))

		if len(tds) > 0 {
			stks = append(stks, stock.Stock{
				Symbol: scrape.Text(tds[0]),
				Name:   scrape.Text(tds[1]),
			})
		}
	}

	var mcs []marketChameleon
	concurrency := 15
	for _, stk := range stks {
		if concurrency == 0 {
			fmt.Printf("Going to rest for %d seconds\n", 15)
			time.Sleep(15 * time.Second)
			fmt.Printf("Waking up after %d seconds sleeping\n", 15)

			concurrency = 15
		}

		concurrency--

		mc := marketChameleon{
			Stock: stk,
		}

		resp, err := http.Get(fmt.Sprintf("https://marketchameleon.com/Overview/%s/Dividends/", stk.Symbol))
		if err != nil {
			panic(err)
		}

		root, err := html.Parse(resp.Body)
		if err != nil {
			panic(err)
		}

		// define a matcher
		matcher := func(n *html.Node) bool {
			// must check for nil values
			if n.DataAtom == atom.Table && scrape.Attr(n, "class") == "mp_lightborder" {
				if n.Parent.DataAtom == atom.Div && n.Parent.Parent.DataAtom == atom.Div && scrape.Attr(n.Parent.Parent, "class") == "symov_div_summary_outer" {
					return true
				}
			}

			return false
		}

		divMarketChameleonTable, ok := scrape.Find(root, matcher)
		if !ok {
			fmt.Printf("Stock %s %s not dividend found\n", stk.Symbol, stk.Name)

			continue
		}

		// define a matcher
		matcher = func(n *html.Node) bool {
			if n.DataAtom == atom.Td && n.Parent.DataAtom == atom.Tr && n.Parent.Parent.DataAtom == atom.Tbody {
				return true
			}

			return false
		}

		tds := scrape.FindAll(divMarketChameleonTable, matcher)

		for i, td := range tds {
			switch i {
			case 0:
				mc.StockDividend.NextExDate = scrape.Text(td)
			case 1:
				mc.StockDividend.Amount = scrape.Text(td)
			case 2:
				mc.StockDividend.ChangePrevYear = scrape.Text(td)
			case 3:
				mc.StockDividend.ForwardYield = strings.Replace(scrape.Text(td), "%", "", 1)
			}
		}

		fmt.Printf("Stock %s %s Forward Yield %s%%\n", mc.Stock.Symbol, mc.Stock.Name, mc.StockDividend.ForwardYield)

		mcs = append(mcs, mc)
	}

	sort.Slice(mcs, func(i, j int) bool { return mcs[i].StockDividend.ForwardYield < mcs[j].StockDividend.ForwardYield })

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)

	wColor := color.New(color.FgWhite).FprintlnFunc()

	wColor(tw, "")
	wColor(tw, "")
	wColor(tw, "# Dividend Growth Stocks")
	wColor(tw, "")
	wColor(tw, "#\t Stocks\t Symbol\t D. Yield")

	for i, mc := range mcs {
		str := fmt.Sprintf(
			"%d\t %s\t %s\t %s%%\t",
			i+1,
			mc.Stock.Name,
			mc.Stock.Symbol,
			mc.StockDividend.ForwardYield,
		)

		wColor(tw, str)
	}

	tw.Flush()

	return nil
}
