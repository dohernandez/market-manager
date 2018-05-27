package cc

import (
	"encoding/json"
	"fmt"
)

const converterUrl = "%s/convert?q=EUR_USD&compact=y"

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

	type ConverterResp struct {
		EURUSD struct {
			Val float64 `json:"val"`
		} `json:"EUR_USD"`
	}
	var cr ConverterResp

	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return c, err
	}

	c.EURUSD = cr.EURUSD.Val

	return c, nil
}
