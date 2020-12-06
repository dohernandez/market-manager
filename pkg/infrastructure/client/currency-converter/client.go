package cc

import (
	"net/http"

	cache "github.com/patrickmn/go-cache"
)

// Client provides methods to interact with http://free.currencyconverterapi.com/ HTTP API for developers.
type Client struct {
	client *http.Client

	baseUrl   string
	Converter *converterEndpoint
}

func NewClient(baseUrl string, apiKey string, client *http.Client, ch *cache.Cache) *Client {
	c := Client{}
	c.baseUrl = baseUrl
	c.client = client

	c.Converter = &converterEndpoint{
		base:   &c,
		cache:  ch,
		apiKey: apiKey,
	}

	return &c
}
