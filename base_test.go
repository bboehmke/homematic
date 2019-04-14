package homematic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/bboehmke/homematic/rpc"
)

type testClient func(method string, params []interface{}) (rpc.Response, error)

func (c testClient) Call(method string, params []interface{}) (rpc.Response, error) {
	return c(method, params)
}


func TestBaseClient_ListMethods(t *testing.T) {
	ass := assert.New(t)

	client := testClient(func(method string, params []interface{}) (rpc.Response, error) {
		ass.Equal("system.listMethods", method)
		ass.Nil(params)

		return rpc.Response{
			Params: []interface{}{
				[]string{"aaa", "bbb"},
			},
		}, nil
	})

	c := BaseClient{client}
	methods, err := c.ListMethods()
	ass.NoError(err)
	ass.Equal([]string{"aaa", "bbb"}, methods)
}