package cc

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const converterUrl = "%s/convert?q=EUR_USD&compact=ultra"

type (
	converterEndpoint struct {
		base *Client
	}

	Converter struct {
		EURUSD float64 `json:"EUR_USD"`
	}
)

func (e *converterEndpoint) Get() (Converter, error) {
	c := Converter{}
	url := fmt.Sprintf(converterUrl, baseUrl)

	resp, err := e.base.client.Get(url)
	if err != nil {
		return c, err
	}

	err = json.NewDecoder(resp.Body).Decode(&c)
	if err != nil {
		return c, err
	}

	c.EURUSD, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", c.EURUSD), 64)

	return c, nil
}
