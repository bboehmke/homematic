package homematic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/bboehmke/homematic/rpc"
	"gitlab.com/bboehmke/homematic/script"
)

type testRpcClient func(method string, params []interface{}) (rpc.Response, error)

func (c testRpcClient) Call(method string, params []interface{}) (rpc.Response, error) {
	return c(method, params)
}

type testScriptClient func(script string) (script.Result, error)

func (c testScriptClient) Call(script string) (script.Result, error) {
	return c(script)
}

func TestBaseClient_ListMethods(t *testing.T) {
	ass := assert.New(t)

	client := testRpcClient(func(method string, params []interface{}) (rpc.Response, error) {
		ass.Equal("system.listMethods", method)
		ass.Nil(params)

		return rpc.Response{
			Params: []interface{}{
				[]string{"aaa", "bbb"},
			},
		}, nil
	})

	c := BaseClient{rpc: client, script: nil}
	methods, err := c.ListMethods()
	ass.NoError(err)
	ass.Equal([]string{"aaa", "bbb"}, methods)
}
