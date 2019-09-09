package script

import (
	"encoding/xml"
	"net/http"
	"strings"
)

// Client interface for remote script client
type Client interface {
	Call(script string) (Result, error)
}

// NewClient creates new client
func NewClient(url string) Client {
	return &client{url, http.DefaultClient}
}

// RPC client
type client struct {
	Url    string
	client *http.Client
}

// Call sends an RPC to server
func (c *client) Call(script string) (Result, error) {
	resp, err := c.client.Post(c.Url+"a.exe", "", strings.NewReader(script))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data Result
	decoder := xml.NewDecoder(resp.Body)
	return data, decoder.Decode(&data)
}
