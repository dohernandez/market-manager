package cc

import "net/http"

const baseUrl = "http://free.currencyconverterapi.com/api/v5"

// Client provides methods to interact with http://free.currencyconverterapi.com/ HTTP API for developers.
type Client struct {
	client *http.Client

	Converter *converterEndpoint
}

func NewClient(client *http.Client) *Client {
	c := Client{}
	c.client = client

	c.Converter = &converterEndpoint{&c}

	return &c
}
