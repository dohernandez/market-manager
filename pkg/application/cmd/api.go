package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/dohernandez/market-manager/pkg/market-manager/purchase/stock"
)

// API ...
type API struct {
	*Base
}

// NewAPI constructs ApiCMD
func NewAPI(baseCMD *Base) *API {
	return &API{
		Base: baseCMD,
	}
}

// Run runs the application import data
func (cmd *API) Run(cliCtx *cli.Context) error {
	return cmd.Run1(cliCtx)
}

// Run runs the application import data
func (cmd *API) Run1(cliCtx *cli.Context) error {
	//ctx, cancelCtx := context.WithCancel(context.TODO())
	//defer cancelCtx()

	type marketChameleon struct {
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

	//
	//// Database connection
	//logger.FromContext(ctx).Info("Initializing database connection")
	//db, err := cmd.initDatabaseConnection()
	//if err != nil {
	//	logger.FromContext(ctx).WithError(err).Fatal("Failed initializing database")
	//}
	//
	//c := cmd.Container(db)
	//
	//stockService := c.PurchaseServiceInstance()
	//stks, err := stockService.Stocks()
	//if err != nil {
	//	fmt.Printf("%+v\n", err)
	//
	//	return err
	//}
	//
	//for _, stk := range stks {
	//	client := iex.NewClient(&http.Client{})
	//
	//	q, err := client.Quote.Get(stk.Symbol)
	//	if err != nil {
	//		fmt.Printf("%+v\n", err)
	//	}
	//
	//	fmt.Printf("%s %+v\n", stk.Symbol, q)
	//}
	//

	///////////////////////////////////////////////////////////////////////////////////////////////////////////
	///////////////////////////////////////////////////////////////////////////////////////////////////////////
	//now := time.Now()
	//wk52back := now.Add(-52 * 7 * 24 * time.Hour)
	//
	//var spy quote.Quote
	//store := cache.New(time.Hour*2, time.Hour*10)
	//
	//key := "etp.52wk"
	//val, found := store.Get(key)
	//fmt.Printf("%+v\n", found)
	//if found {
	//	var ok bool
	//	if spy, ok = val.(quote.Quote); !ok {
	//		return errors.New("cache value invalid for Quote")
	//	}
	//} else {
	//	spy, _ = quote.NewQuoteFromYahoo("etp", wk52back.Format("2006-01-02"), now.Format("2006-01-02"), quote.Daily, true)
	//	fmt.Print(spy.CSV())
	//
	//	store.Set(key, spy, cache.DefaultExpiration)
	//
	//	_, found := store.Get(key)
	//	fmt.Printf("%+v\n", found)
	//}
	//
	//high52wk := spy.High[0]
	//low52wk := spy.Low[0]
	//for k := range spy.Date[1:] {
	//	if high52wk < spy.High[k] {
	//		high52wk = spy.High[k]
	//	}
	//
	//	if low52wk > spy.Low[k] {
	//		low52wk = spy.Low[k]
	//	}
	//}
	//
	//fmt.Printf("52 wk start: %s  end %s high [%.2f] - low [%.2f]\n", wk52back.Format("2006-01-02"), now.Format("2006-01-02"), high52wk, low52wk)
	//
	//
	//prices, err := gf.GetPrices(ctx, &gf.Query{P: "1Y", I: "86400", X: "NYSE", Q: "ETP"})
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println(prices)

	///////////////////////////////////////////////////////////////////////////////////////////////////////////
	// SCRAPE
	///////////////////////////////////////////////////////////////////////////////////////////////////////////

	//
	//resp, err = http.Get(fmt.Sprintf("https://marketchameleon.com/Overview/%s/", "CSCO"))
	//if err != nil {
	//	panic(err)
	//}
	//
	//root, err = html.Parse(resp.Body)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// define a matcher
	//matcher = func(n *html.Node) bool {
	//	// must check for nil values
	//	if n.DataAtom == atom.Div && scrape.Attr(n, "class") == "symov_stat_box symov_info_box _c" {
	//		return true
	//	}
	//
	//	return false
	//}
	//
	//divMarketChameleonInfoDiv, ok := scrape.Find(root, matcher)
	//if !ok {
	//	return errors.New("Stock info not found")
	//}
	//
	//divs := scrape.FindAll(divMarketChameleonInfoDiv, scrape.ByClass("flex_container_between"))
	//
	//for i, div := range divs {
	//	d, ok := scrape.Find(div, scrape.ByClass("datatag"))
	//	if !ok {
	//		continue
	//	}
	//
	//	switch i {
	//	case 0:
	//		mc.StockInfo.Type = scrape.Text(d)
	//	case 1:
	//		mc.StockInfo.Sector = scrape.Text(d)
	//	case 2:
	//		mc.StockInfo.Industry = scrape.Text(d)
	//	}
	//}
	//
	//// define a matcher
	//matcher = func(n *html.Node) bool {
	//	if n.DataAtom == atom.Div && scrape.Attr(n, "class") == "symov_stat_box _c" {
	//		if n.FirstChild != nil && n.FirstChild.NextSibling != nil && scrape.Text(n.FirstChild.NextSibling) == "Volatility" {
	//			return true
	//		}
	//	}
	//
	//	return false
	//}
	//
	//divMarketChameleonVolatilityDiv, ok := scrape.Find(root, matcher)
	//if !ok {
	//	return errors.New("Stock volatility not found")
	//}
	//
	//divs = scrape.FindAll(divMarketChameleonVolatilityDiv, scrape.ByClass("flex_container_between"))
	//
	//for i, div := range divs {
	//	d, ok := scrape.Find(div, scrape.ByClass("datatag"))
	//	if !ok {
	//		continue
	//	}
	//
	//	switch i {
	//	case 1:
	//		mc.StockVolatility.HV20Day = scrape.Text(d)
	//	case 2:
	//		mc.StockVolatility.HV52Week = scrape.Text(d)
	//	}
	//}
	//
	//fmt.Printf("%+v\n", mc)

	//stk := stock.Stock{Symbol: "AXP"}

	//ps := service.NewYahooScrapeStockPrice(cmd.ctx, "https://finance.yahoo.com/quote")
	//
	//p, err := ps.Price(&stk)
	//if err != nil {
	//	fmt.Printf("%+v", err)
	//
	//	return err
	//}
	//
	//fmt.Printf("%+v\n", p)
	//
	//ss := service.NewStockSummaryMarketChameleon(cmd.ctx, "https://marketchameleon.com/Overview")
	//
	//s, err := ss.Summary(&stk)
	//if err != nil {
	//	fmt.Printf("%+v", err)
	//
	//	return err
	//}
	//
	//fmt.Printf("%+v\n", s)
	//
	//pvs := service.NewMarketChameleonStockPriceVolatility(cmd.ctx, "https://marketchameleon.com/Overview")
	//
	//pv, err := pvs.PriceVolatility(&stk)
	//if err != nil {
	//	fmt.Printf("%+v", err)
	//
	//	return err
	//}
	//
	////fmt.Printf("%+v\n", pv)
	//
	//stk := stock.Stock{Symbol: "AXP"}
	//
	//pvs := service.NewStockDividendMarketChameleon(cmd.ctx, "https://marketchameleon.com/Overview")
	////shd, _ := time.Parse("2-Jan-2006", "1-Jan-2017")
	//
	//pv, err := pvs.NextFuture(&stk)
	//if err != nil {
	//	fmt.Printf("%+v", err)
	//
	//	return err
	//}
	//
	//fmt.Printf("%+v\n", pv)

	return nil
}
