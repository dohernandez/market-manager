package cc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/patrickmn/go-cache"
)

const converterUrl = "%s/convert?q=%s&compact=ultra"

type (
	converterEndpoint struct {
		base  *Client
		cache *cache.Cache
	}

	Converter struct {
		EURUSD float64 `json:"EUR_USD"`
		EURCAD float64 `json:"EUR_CAD"`
	}
)

func (e *converterEndpoint) Get() (Converter, error) {
	c := Converter{}
	key := "Converter"

	val, found := e.cache.Get(key)
	if found {
		c, ok := val.(Converter)
		if !ok {
			return c, errors.New("cache value invalid for Converter")
		}

		return c, nil
	}

	err := e.eurUSD(&c)
	if err != nil {
		return c, err
	}

	err = e.eurCAD(&c)
	if err != nil {
		return c, err
	}

	e.cache.Set(key, c, cache.DefaultExpiration)

	return c, nil
}

func (e *converterEndpoint) eurUSD(c *Converter) error {
	url := fmt.Sprintf(converterUrl, baseUrl, "EUR_USD")

	resp, err := e.base.client.Get(url)
	if err != nil {
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(c)
	if err != nil {
		return err
	}

	c.EURUSD, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", c.EURUSD), 64)

	return nil
}

func (e *converterEndpoint) eurCAD(c *Converter) error {
	url := fmt.Sprintf(converterUrl, baseUrl, "EUR_CAD")

	resp, err := e.base.client.Get(url)
	if err != nil {
		return err
	}

	err = json.NewDecoder(resp.Body).Decode(c)
	if err != nil {
		return err
	}

	c.EURCAD, _ = strconv.ParseFloat(fmt.Sprintf("%.4f", c.EURCAD), 64)

	return nil
}
