package command

import (
	"context"
	"fmt"

	"github.com/urfave/cli"

	gf "github.com/dohernandez/googlefinance-client-go"
)

// ApiCommand ...
type ApiCommand struct {
	*BaseCommand
}

// NewApiCommand constructs ApiCommand
func NewApiCommand(baseCommand *BaseCommand) *ApiCommand {
	return &ApiCommand{
		BaseCommand: baseCommand,
	}
}

// Run runs the application import data
func (cmd *ApiCommand) Run(cliCtx *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.TODO())
	defer cancelCtx()
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

	prices, err := gf.GetPrices(ctx, &gf.Query{P: "1Y", I: "86400", X: "NYSE", Q: "ETP"})
	if err != nil {
		panic(err)
	}

	fmt.Println(prices)

	return nil
}
