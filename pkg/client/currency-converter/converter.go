package cc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/patrickmn/go-cache"
)

const converterUrl = "%s/convert?q=EUR_USD&compact=ultra"

type (
	converterEndpoint struct {
		base  *Client
		store *cache.Cache
	}

	Converter struct {
		EURUSD float64 `json:"EUR_USD"`
	}
)

func (e *converterEndpoint) Get() (Converter, error) {
	c := Converter{}
	key := "Converter.EUR_USD"

	val, found := e.store.Get(key)
	if found {
		c, ok := val.(Converter)
		if !ok {
			return c, errors.New("cache value invalid for Converter")
		}

		return c, nil
	}

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

	e.store.Set(key, c, cache.DefaultExpiration)

	return c, nil
}
