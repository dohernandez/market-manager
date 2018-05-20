package iex

import (
	"encoding/json"
	"fmt"
)

const quoteUrl = "%s/stock/%s/quote"

type (
	quoteEndpoint struct {
		base *Client
	}

	Quote struct {
		Symbol       string  `json:"symbol"`
		CompanyName  string  `json:"companyName"`
		Open         float64 `json:"open"`
		Close        float64 `json:"close"`
		LatestUpdate int     `json:"latestUpdate"`
	}
)

func (e *quoteEndpoint) Get(symbol string) (Quote, error) {
	q := Quote{}
	url := fmt.Sprintf(quoteUrl, baseUrl, symbol)

	res, err := e.base.client.Get(url)
	if err != nil {
		return q, err
	}

	err = json.NewDecoder(res.Body).Decode(&q)
	if err != nil {
		return q, err
	}

	return q, nil
}
