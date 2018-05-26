package iex

import (
	"encoding/json"
	"fmt"
	"time"
)

const quoteUrl = "%s/stock/%s/quote"

type (
	quoteEndpoint struct {
		base *Client
	}

	Quote struct {
		LatestUpdate time.Time `json:"latestUpdate"`
		Symbol       string    `json:"symbol"`
		CompanyName  string    `json:"companyName"`
		Close        float64   `json:"close"`
		High         float64   `json:"high"`
		Low          float64   `json:"low"`
		Open         float64   `json:"open"`
		Volume       float64   `json:"volume"`
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
