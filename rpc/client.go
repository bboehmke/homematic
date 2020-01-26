package rpc

import (
	"bytes"
	"encoding/xml"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client interface for XML RPC client
type Client interface {
	Call(method string, params []interface{}) (*Response, error)
	LocalIP() (string, error)
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
func (c *client) Call(method string, params []interface{}) (*Response, error) {
	buf := new(bytes.Buffer)
	err := xml.NewEncoder(buf).Encode(Request{
		Method: method,
		Params: params,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.URL, "text/xml", buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ParseResponse(resp.Body)
}

func (c *client) LocalIP() (string, error) {
	// get host from url
	u, err := url.Parse(c.URL)
	if err != nil {
		return "", err
	}

	// create TCP connection
	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// get local address of connection
	addr := conn.LocalAddr()
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		return "", errors.New("invalid")
	}
	return tcpAddr.IP.String(), nil
}
