package homematic

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/bboehmke/homematic/rpc"
)

func TestDevice_nameChanged(t *testing.T) {
	ass := assert.New(t)

	device := &Device{
		Name: "aaa",
	}

	ass.Equal("aaa", device.Name)
	device.nameChanged("bbb")
	ass.Equal("bbb", device.Name)
}

func TestDevice_valueChanged(t *testing.T) {
	ass := assert.New(t)

	device := new(Device)

	device.SetValueChangedHandler(func(key string, value interface{}) {
		ass.Equal("aaa", key)
		ass.Equal("bbb", value)
	})
	device.valueChanged("aaa", "bbb")
}

func TestDevice_HasValues(t *testing.T) {
	ass := assert.New(t)

	device := new(Device)

	ass.False(device.HasValues())
	device.ParamSets = []string{"VALUES"}
	ass.True(device.HasValues())
}

func TestDevice_GetValues(t *testing.T) {
	ass := assert.New(t)

	device := &Device{
		Address: "address",
	}

	device.client = testRpcClient(func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("getParamset", method)
		ass.Equal([]interface{}{"address", "VALUES"}, params)
		return nil, errors.New("test")
	})
	_, err := device.GetValues()
	ass.EqualError(err, "test")

	device.client = testRpcClient(func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("getParamset", method)
		ass.Equal([]interface{}{"address", "VALUES"}, params)
		return &rpc.Response{
			Params: []interface{}{
				map[string]interface{}{
					"aaa": 111,
					"bbb": 222,
				},
			},
		}, nil
	})
	values, err := device.GetValues()
	ass.NoError(err)
	ass.Equal(map[string]interface{}{
		"aaa": 111,
		"bbb": 222,
	}, values)
}

func TestDevice_GetValue(t *testing.T) {
	ass := assert.New(t)

	device := &Device{
		Address: "address",
	}

	device.client = testRpcClient(func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("getValue", method)
		ass.Equal([]interface{}{"address", "testDevice"}, params)
		return nil, errors.New("test")
	})
	_, err := device.GetValue("testDevice")
	ass.EqualError(err, "test")

	device.client = testRpcClient(func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("getValue", method)
		ass.Equal([]interface{}{"address", "testDevice"}, params)
		return &rpc.Response{
			Params: []interface{}{
				111,
			},
		}, nil
	})
	value, err := device.GetValue("testDevice")
	ass.NoError(err)
	ass.Equal(111, value)
}

func TestDevice_SetValue(t *testing.T) {
	ass := assert.New(t)

	device := &Device{
		Address: "address",
	}

	device.client = testRpcClient(func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("setValue", method)
		ass.Equal([]interface{}{"address", "aaa", 111}, params)
		return nil, errors.New("test")
	})
	err := device.SetValue("aaa", 111)
	ass.EqualError(err, "test")
}

func TestDevice_GetValuesDescription(t *testing.T) {
	ass := assert.New(t)

	device := &Device{
		Address: "address",
	}

	device.client = testRpcClient(func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("getParamsetDescription", method)
		ass.Equal([]interface{}{"address", "VALUES"}, params)
		return &rpc.Response{
			Params: []interface{}{
				map[string]interface{}{
					"aaa": map[string]interface{}{
						"ID":         "aaa",
						"DEFAULT":    123,
						"TYPE":       "type",
						"UNIT":       "unit",
						"TAB_ORDER":  3,
						"OPERATIONS": 0x01 + 0x02 + 0x04,
						"FLAGS":      0x01 + 0x02 + 0x04 + 0x08 + 0x10,
						"VALUE_LIST": []interface{}{
							"aaa", "bbb",
						},
					},
				},
			},
		}, nil
	})
	values, err := device.GetValuesDescription()
	ass.NoError(err)
	ass.Equal(map[string]ParameterDescription{
		"aaa": {
			ID:       "aaa",
			Default:  123,
			Type:     "type",
			Unit:     "unit",
			TabOrder: 3,
			ValueList: []string{
				"aaa", "bbb",
			},

			OperationRead:  true,
			OperationWrite: true,
			OperationEvent: true,

			FlagVisible:   true,
			FlagInternal:  true,
			FlagTransform: true,
			FlagService:   true,
			FlagSticky:    true,
		},
	}, values)

}
