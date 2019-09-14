package rpc

import (
	"bytes"
	"encoding/xml"
	"net/http"
)

// Client interface for XML RPC client
type Client interface {
	Call(method string, params []interface{}) (*Response, error)
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
func (c *client) Call(method string, params []interface{}) (*Response, error) {
	buf := new(bytes.Buffer)
	err := xml.NewEncoder(buf).Encode(Request{
		Method: method,
		Params: params,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.Url, "text/xml", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ParseResponse(resp.Body)
}
