package homematic

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.com/bboehmke/homematic/rpc"
	"gitlab.com/bboehmke/homematic/script"
)

type testRpcClient func(method string, params []interface{}) (*rpc.Response, error)

func (c testRpcClient) Call(method string, params []interface{}) (*rpc.Response, error) {
	return c(method, params)
}

func (c testRpcClient) LocalIP() (string, error) {
	return "127.0.0.1", nil
}

type testScriptClient func(script string) (script.Result, error)

func (c testScriptClient) Call(script string) (script.Result, error) {
	return c(script)
}

func TestCCU_checkEventHandling(t *testing.T) {
	ass := assert.New(t)

	ccu, err := NewCCU("127.0.0.1")
	ass.NoError(err)

	var client testRpcClient = func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("init", method)
		ass.Equal([]interface{}{
			fmt.Sprintf("http://127.0.0.1:%d", ccu.rpcServer.Port()),
			"test",
		}, params)
		return &rpc.Response{}, nil
	}
	ccu.rpcClients = map[string]rpc.Client{
		"test": client,
	}

	ass.NoError(ccu.Start())

	ccu.lastClientEvent = make(map[string]time.Time)

	ass.NoError(ccu.checkEventHandling())

	client = func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("init", method)
		ass.Equal([]interface{}{
			fmt.Sprintf("http://127.0.0.1:%d", ccu.rpcServer.Port()),
			"",
		}, params)
		return &rpc.Response{}, nil
	}
	ccu.rpcClients = map[string]rpc.Client{
		"test": client,
	}
	ass.NoError(ccu.Stop())
}

func TestCCU_GetDevices(t *testing.T) {
	ass := assert.New(t)

	ccu, err := NewCCU("127.0.0.1")
	ass.NoError(err)

	var rpcClient testRpcClient = func(method string, params []interface{}) (*rpc.Response, error) {
		ass.Equal("listDevices", method)
		ass.Nil(params)
		return &rpc.Response{
			Params: []interface{}{
				[]interface{}{
					map[string]interface{}{
						"ADDRESS": "address",
						"TYPE":    "switch",
					},
				},
			},
		}, nil
	}
	ccu.rpcClients = map[string]rpc.Client{
		"test": rpcClient,
	}

	var scriptClient testScriptClient = func(script string) (script.Result, error) {
		return map[string]string{
			"output": "address=testDevice\naddress2=testDevice2\n",
		}, nil
	}
	ccu.scriptClient = scriptClient

	devices, err := ccu.GetDevices()
	ass.NoError(err)
	ass.Equal(map[string]*Device{
		"address": ccu.devices["address"],
	}, devices)
	ass.Equal("testDevice", ccu.devices["address"].Name)
}
