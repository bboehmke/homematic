package homematic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/bboehmke/homematic/rpc"
)

func TestBaseClient_ListDevices(t *testing.T) {
	ass := assert.New(t)

	client := testClient(func(method string, params []interface{}) (rpc.Response, error) {
		ass.Equal("listDevices", method)
		ass.Nil(params)

		return rpc.Response{
			Params: []interface{}{
				[]interface{}{
					map[string]interface{}{
						"TYPE": "aaa",
						"ADDRESS": "bbb",
						"CHILDREN": []string{"a", "b"},
						"PARENT": "c",
						"PARAMSETS": []string{"VALUES", "EVENTS"},
					},
					map[string]interface{}{
						"TYPE": "111",
						"ADDRESS": "222",
						"CHILDREN": []string{"1", "2"},
						"PARENT": "3",
						"PARAMSETS": []string{"VALUES", "EVENTS"},
					},
				},
			},
		}, nil
	})

	c := BaseClient{client}
	result, err := c.ListDevices()
	ass.NoError(err)
	ass.Equal([]DeviceDescription{{
		Type:"aaa",
		Address:"bbb",
		Children:[]string{"a", "b"},
		Parent:"c",
		ParamSets:[]string{"VALUES", "EVENTS"},
	}, {
		Type:"111",
		Address:"222",
		Children:[]string{"1", "2"},
		Parent:"3",
		ParamSets:[]string{"VALUES", "EVENTS"},
	}}, result)
}

func TestBaseClient_GetValues(t *testing.T) {
	ass := assert.New(t)

	client := testClient(func(method string, params []interface{}) (rpc.Response, error) {
		ass.Equal("getParamset", method)
		ass.Equal([]interface{}{
			"aaa", "VALUES",
		},params)

		return rpc.Response{
			Params: []interface{}{
				map[string]interface{}{
					"STATE": "aaa",
					"ADDRESS": 42,
				},
			},
		}, nil
	})

	c := BaseClient{client}
	result, err := c.GetValues("aaa")
	ass.NoError(err)
	ass.Equal(map[string]interface{}{
		"STATE": "aaa",
		"ADDRESS": 42,
	}, result)
}

func TestBaseClient_GetValue(t *testing.T) {
	ass := assert.New(t)

	client := testClient(func(method string, params []interface{}) (rpc.Response, error) {
		ass.Equal("getValue", method)
		ass.Equal([]interface{}{
			"aaa", "bbb",
		},params)

		return rpc.Response{
			Params: []interface{}{
				42,
			},
		}, nil
	})

	c := BaseClient{client}
	result, err := c.GetValue("aaa", "bbb")
	ass.NoError(err)
	ass.Equal(42, result)
}

func TestBaseClient_SetValue(t *testing.T) {
	ass := assert.New(t)

	client := testClient(func(method string, params []interface{}) (rpc.Response, error) {
		ass.Equal("setValue", method)
		ass.Equal([]interface{}{
			"aaa", "bbb", 42,
		},params)

		return rpc.Response{}, nil
	})

	c := BaseClient{client}
	ass.NoError(c.SetValue("aaa", "bbb", 42))
}