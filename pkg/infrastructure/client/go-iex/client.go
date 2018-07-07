package iex

import "net/http"

const baseUrl = "https://api.iextrading.com/1.0"

// Client provides methods to interact with IEX's HTTP API for developers.
type Client struct {
	client *http.Client

	Quote *quoteEndpoint
}

func NewClient(client *http.Client) *Client {
	c := Client{}
	c.client = client

	c.Quote = &quoteEndpoint{&c}

	return &c
}
