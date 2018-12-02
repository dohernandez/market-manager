package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"time"

	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func main() {
	// create context
	ctxt, cancel := context.WithCancel(context.Background())

	c, err := chromedp.New(ctxt, chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		log.Println("main defer")
		shutdown(ctxt, c, cancel)
	}()

	handleGracefulShutdown(ctxt, c, cancel)

	var stocks []string
	for _, s := range []string{
		"LLY",
		"CSCO",
		"VZ",
		"SBUX",
		"V",
	} {
		// run task list
		res, err := runTaskList(ctxt, c, s)
		if err != nil {
			log.Println(err)
		}

		stocks = append(stocks, fmt.Sprintf("\"%s\": %s", s, res.(string)))

		time.Sleep(5 * time.Second)
	}

	log.Printf("overview: {%s}", strings.Join(stocks, ", "))
}

func handleGracefulShutdown(ctxt context.Context, c *chromedp.CDP, cancel func()) {
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGTERM, os.Interrupt)

	go func() {
		<-exit

		gracefulDelay := 20 * time.Second
		// Killing with no mercy after a graceful delay
		go func() {
			time.Sleep(gracefulDelay)

			log.Fatalf("Failed to gracefully shutdown server in %s, exiting.", gracefulDelay)

			os.Exit(0)
		}()

		log.Println("Graceful shutdown")
		shutdown(ctxt, c, cancel)
	}()
}

func shutdown(ctxt context.Context, c *chromedp.CDP, cancel func()) {
	log.Println("Shutting down ...")

	// shutdown chrome
	err := c.Shutdown(ctxt)
	if err != nil {
		log.Fatal(err)
	}

	// wait for chrome to finish
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}

	if cancel != nil {
		cancel()
	}
}

const textJS = `(function(a) {
		var s = '';
		for (var i = 0; i < a.length; i++) {
			var current = a[i];
			
			// Looping over tbody
			if (current.offsetParent !== null && current.nodeName == 'TBODY') {

				// Check the TBODY has children
				if(current.children && current.children.length > 0) {
					var dict = [];

					for(var j = 0; j < current.children.length; j++) {
						var child = current.children[j];

						// Check the TR has children
						if(child.children && child.children.length > 0) {
							
							dict.push({
    							ex_dates: child.children[0].textContent,
								r_date: child.children[1].textContent,
								p_date: child.children[2].textContent,
								status: child.children[3].textContent,
								amount: child.children[5].textContent
							});
						}
					}
					s += JSON.stringify(dict);
				}
			}
		}

		return s;
	})($x('%s/node()'))`

func runTaskList(ctxt context.Context, c *chromedp.CDP, stock string) (interface{}, error) {
	ctxtRun, cancelRun := context.WithTimeout(ctxt, 25*time.Second)
	defer func() {
		log.Println("runTaskList defer")
		cancelRun()
	}()

	//var res []*cdp.Node
	var res string
	sel := `future_divs`

	err := c.Run(ctxtRun, chromedp.Tasks{
		chromedp.Navigate(fmt.Sprintf(`https://marketchameleon.com/Overview/%s/Dividends/`, stock)),
		chromedp.Sleep(2 * time.Second),
		chromedp.WaitReady(sel, chromedp.ByID),
		chromedp.QueryAfter(sel, func(ctxt context.Context, h *chromedp.TargetHandler, nodes ...*cdp.Node) error {
			if len(nodes) < 1 {
				return fmt.Errorf("selector `%s` did not return any nodes", sel)
			}

			return chromedp.EvaluateAsDevTools(fmt.Sprintf(textJS, nodes[0].FullXPath()), &res).Do(ctxt, h)
		}),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}
