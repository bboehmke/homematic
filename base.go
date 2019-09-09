package homematic

import (
	"fmt"

	"github.com/spf13/cast"

	"gitlab.com/bboehmke/homematic/rpc"
	"gitlab.com/bboehmke/homematic/script"
)

// NewClient creates new client to a homematic CCU (RF & wired)
func NewClient(host string) *Client {
	s := script.NewClient(fmt.Sprintf("http://%s:8181/", host))
	return &Client{
		Wired: &BaseClient{
			rpc:    rpc.NewClient(fmt.Sprintf("http://%s:2000/", host)),
			script: s,
		},
		RF: &BaseClient{
			rpc:    rpc.NewClient(fmt.Sprintf("http://%s:2001/", host)),
			script: s,
		},
	}
}

// Client holds wired and RF clients
type Client struct {
	Wired *BaseClient
	RF    *BaseClient
}

// BaseClient provides functionality to interact with CCU
type BaseClient struct {
	rpc    rpc.Client
	script script.Client
}

// ListMethods returns a list with available methods
func (c *BaseClient) ListMethods() ([]string, error) {
	response, err := c.rpc.Call("system.listMethods", nil)
	if err != nil {
		return nil, err
	}
	return cast.ToStringSlice(response.FirstParam()), nil
}
