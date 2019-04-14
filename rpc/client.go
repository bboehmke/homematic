package rpc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"

	"golang.org/x/net/html/charset"
)

// Client interface for XML RPC client
type Client interface {
	Call(method string, params []interface{}) (Response, error)
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
func (c *client) Call(method string, params []interface{}) (Response, error) {
	buf := new(bytes.Buffer)
	err := xml.NewEncoder(buf).Encode(Request{
		Method: method,
		Params: params,
	})
	if err != nil {
		return Response{}, nil
	}

	resp, err := c.client.Post(c.Url, "text/xml", buf)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	var response Response
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&response)

	if response.Fault != nil {
		return response, fmt.Errorf("RPC error %s", response.Fault.String)
	}

	return response, nil
}
