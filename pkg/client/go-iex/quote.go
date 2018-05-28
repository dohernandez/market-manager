package iex

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dohernandez/market-manager/pkg/market-manager"
)

const quoteUrl = "%s/stock/%s/quote"

type (
	quoteEndpoint struct {
		base *Client
	}

	Quote struct {
		LatestUpdate int     `json:"latestUpdate"`
		Symbol       string  `json:"symbol"`
		CompanyName  string  `json:"companyName"`
		Close        float64 `json:"close"`
		High         float64 `json:"high"`
		Low          float64 `json:"low"`
		Open         float64 `json:"open"`
		Volume       float64 `json:"volume"`
	}
)

func (e *quoteEndpoint) Get(symbol string) (Quote, error) {
	q := Quote{}
	url := fmt.Sprintf(quoteUrl, baseUrl, symbol)

	resp, err := e.base.client.Get(url)
	if err != nil {
		return q, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return q, mm.ErrNotFound
	}

	err = json.NewDecoder(resp.Body).Decode(&q)
	if err != nil {
		fmt.Printf("%+v", resp)
		return q, err
	}

	return q, nil
}
