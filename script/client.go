package script

import (
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

// Client interface for remote script client
type Client interface {
	Call(script string) (Result, error)
}

// NewClient creates new client
func NewClient(url string) Client {
	return &client{url, &http.Client{
		Timeout: time.Second * 5,
	}}
}

// RPC client
type client struct {
	URL    string
	client *http.Client
}

// Call sends an RPC to server
func (c *client) Call(script string) (Result, error) {
	resp, err := c.client.Post(c.URL+"a.exe", "", strings.NewReader(script))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// no processing instruction in response
	// -> charset.NewReaderLabel not working
	// -> manually decode iso-8859-1
	var data Result
	decoder := xml.NewDecoder(charmap.Windows1252.NewDecoder().Reader(resp.Body))
	return data, decoder.Decode(&data)
}
