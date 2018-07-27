package bootstrap

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/jmoiron/sqlx"
)

type stockCommandContext struct {
	db         *sqlx.DB
	stocksInfo map[string]string
	stocks     map[string]string

	yahooFinanceWeb  *WireMock
	query1FinanceWeb *WireMock
}

func RegisterStockCommandContext(s *godog.Suite, db *sqlx.DB, yahooFinanceWeb, query1FinanceWeb *WireMock) {
	scc := &stockCommandContext{
		db:               db,
		stocksInfo:       map[string]string{},
		stocks:           map[string]string{},
		yahooFinanceWeb:  yahooFinanceWeb,
		query1FinanceWeb: query1FinanceWeb,
	}

	s.Step(`^following stocks info should be stored:$`, scc.followingStocksInfoShouldBeStored)
	s.Step(`^following stocks should be stored:$`, scc.followingStocksShouldBeStored)
	s.Step(`^that we get the following stocks price from yahoo finance:$`, scc.thatWeGetTheFollowingStocksPriceFromYahooFinance)
}

func (c *stockCommandContext) followingStocksInfoShouldBeStored(stockInfos *gherkin.DataTable) error {
	query := `SELECT id FROM stock_info WHERE name = $1 AND type  = $2`

	for _, row := range stockInfos.Rows[1:] {
		var id string

		err := c.db.Get(&id, query, row.Cells[1].Value, row.Cells[2].Value)
		if err != nil {
			return err
		}

		c.stocksInfo[row.Cells[0].Value] = id
	}

	return nil
}

func (c *stockCommandContext) followingStocksShouldBeStored(stocks *gherkin.DataTable) error {
	query := `SELECT s.id FROM stock s`
	var innerJoin []string
	var where []string

	var existsID = regexp.MustCompile(`^id_(.*)$`)

	for i, cell := range stocks.Rows[0].Cells[1:] {
		var existsExchange = regexp.MustCompile(`^exchange_(.*)$`)
		if match := existsExchange.FindStringSubmatch(cell.Value); len(match) > 0 {
			innerJoin = append(innerJoin, "INNER JOIN exchange e ON s.exchange_id  = e.id")

			where = append(where, fmt.Sprintf("e.%s=$%d", match[1], i+1))

			continue
		}

		if match := existsID.FindStringSubmatch(cell.Value); len(match) > 0 {
			where = append(where, fmt.Sprintf("s.%s=$%d", match[1], i+1))

			continue
		}

		where = append(where, fmt.Sprintf("s.%s=$%d", cell.Value, i+1))
	}

	if len(innerJoin) > 0 {
		query = fmt.Sprintf("%s %s", query, strings.Join(innerJoin, " "))
	}

	if len(where) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(where, " AND "))
	}

	for _, row := range stocks.Rows[1:] {
		var id string
		var args []interface{}

		for k, cell := range row.Cells[1:] {
			if existsID.MatchString(stocks.Rows[0].Cells[k+1].Value) {
				args = append(args, c.stocksInfo[cell.Value])

				continue
			}

			args = append(args, cell.Value)
		}

		err := c.db.Get(&id, query, args...)
		if err != nil {
			return err
		}

		c.stocks[row.Cells[0].Value] = id
	}

	return nil
}

func (c *stockCommandContext) thatWeGetTheFollowingStocksPriceFromYahooFinance(stocksPrice *gherkin.DataTable) error {
	crumb := "H0Im79Q2Cwh"

	//yahoo finance first call mocked
	wmRequest := NewWireMockRequest()
	wmRequest.SetMethod(http.MethodGet)
	wmRequest.SetURL("/")
	c.yahooFinanceWeb.SetWireMockRequest(wmRequest)

	wmResponse := NewWireMockResponse()
	wmResponse.SetHeader("Content-Type", "text/plain;charset=utf-8")
	wmResponse.SetStatus(http.StatusOK)
	c.yahooFinanceWeb.SetWireMockResponse(wmResponse)

	err := c.yahooFinanceWeb.Send()
	if err != nil {
		return err
	}

	// yahoo finance second call mocked
	wmRequest = NewWireMockRequest()
	wmRequest.SetMethod(http.MethodGet)
	wmRequest.SetURL("/v1/test/getcrumb")
	c.query1FinanceWeb.SetWireMockRequest(wmRequest)

	wmResponse = NewWireMockResponse()
	wmResponse.SetTextBody(crumb)
	wmResponse.SetStatus(http.StatusOK)
	c.query1FinanceWeb.SetWireMockResponse(wmResponse)

	err = c.query1FinanceWeb.Send()
	if err != nil {
		return err
	}

	//yahoo finance third call mocked
	endDate := time.Now()
	startDate := endDate.Add(-24 * time.Hour)

	from, _ := time.Parse("2006-01-02", startDate.Format("2006-01-02"))
	to, _ := time.Parse("2006-01-02", endDate.Format("2006-01-02"))

	for _, sp := range stocksPrice.Rows[1:] {
		wmRequest = NewWireMockRequest()
		wmRequest.SetMethod(http.MethodGet)
		wmRequest.SetURL(fmt.Sprintf(
			"/v7/finance/download/%s?period1=%d&period2=%d&interval=1d&events=history&crumb=%s",
			sp.Cells[0].Value,
			from.Unix(),
			to.Unix(),
			crumb))
		c.query1FinanceWeb.SetWireMockRequest(wmRequest)

		body := fmt.Sprintf("%s\n%s",
			"Date,Open,High,Low,Close,Adj Close,Volume",
			fmt.Sprintf(
				"%s,%s,%s,%s,%s,%s,%s",
				sp.Cells[1].Value,
				sp.Cells[2].Value,
				sp.Cells[3].Value,
				sp.Cells[4].Value,
				sp.Cells[5].Value,
				sp.Cells[6].Value,
				sp.Cells[7].Value,
			),
		)
		if err != nil {
			return err
		}

		wmResponse = NewWireMockResponse()
		wmResponse.SetTextBody(body)
		wmResponse.SetStatus(http.StatusOK)
		c.query1FinanceWeb.SetWireMockResponse(wmResponse)

		err = c.query1FinanceWeb.Send()
		if err != nil {
			return err
		}
	}

	return nil
}
