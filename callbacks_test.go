package homematic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/bboehmke/homematic/rpc"
	"gitlab.com/bboehmke/homematic/script"
)

func TestCCU_handleCallback(t *testing.T) {
	ass := assert.New(t)

	ccu, err := NewCCU("127.0.0.1")
	ass.NoError(err)

	resp, fault := ccu.handleCallback("unknown", nil)
	ass.Equal([]interface{}{true}, resp)
	ass.Nil(fault)

	resp, fault = ccu.handleCallback("event", nil)
	ass.Nil(resp)
	ass.Equal(&rpc.Fault{
		Code:   -1,
		String: "invalid event call",
	}, fault)

	resp, fault = ccu.handleCallback("listDevices", nil)
	ass.Equal([]interface{}{[]interface{}{}}, resp)
	ass.Nil(fault)

	resp, fault = ccu.handleCallback("newDevices", nil)
	ass.Nil(resp)
	ass.Equal(&rpc.Fault{
		Code:   -1,
		String: "invalid newDevices call",
	}, fault)
}

func TestCCU_callbackEvent(t *testing.T) {
	ass := assert.New(t)

	ccu, err := NewCCU("127.0.0.1")
	ass.NoError(err)

	resp, fault := ccu.callbackEvent(nil)
	ass.Nil(resp)
	ass.Equal(&rpc.Fault{
		Code:   -1,
		String: "invalid event call",
	}, fault)

	resp, fault = ccu.callbackEvent([]interface{}{
		"id", "address", "name", "value",
	})
	ass.Nil(resp)
	ass.Nil(fault)

	ccu.devices["address"] = new(Device)
	resp, fault = ccu.callbackEvent([]interface{}{
		"id", "address", "name", "value",
	})
	ass.Nil(resp)
	ass.Nil(fault)
}

func TestCCU_callbackListDevices(t *testing.T) {
	ass := assert.New(t)

	ccu, err := NewCCU("127.0.0.1")
	ass.NoError(err)

	ccu.devices["aaa"] = &Device{
		Address: "bb",
		Version: 42,
	}
	resp, fault := ccu.callbackListDevices()
	ass.Equal([]interface{}{[]interface{}{
		map[string]interface{}{
			"ADDRESS": "bb",
			"VERSION": 42,
		},
	}}, resp)
	ass.Nil(fault)
}

func TestCCU_callbackNewDevices(t *testing.T) {
	ass := assert.New(t)

	ccu, err := NewCCU("127.0.0.1")
	ass.NoError(err)

	resp, fault := ccu.callbackNewDevices([]interface{}{"unknown", nil})
	ass.Nil(resp)
	ass.Equal(&rpc.Fault{
		Code:   -1,
		String: "invalid interface id",
	}, fault)

	resp, fault = ccu.callbackNewDevices([]interface{}{
		"go-rf",
		[]interface{}{
			map[string]interface{}{
				"ADDRESS": "address",
				"TYPE":    "switch",
			},
		},
	})
	ass.Equal([]interface{}{true}, resp)
	ass.Nil(fault)

	var scriptClient testScriptClient = func(script string) (script.Result, error) {
		return map[string]string{
			"output": "address=testDevice\naddress2=testDevice2\n",
		}, nil
	}
	ccu.scriptClient = scriptClient
	resp, fault = ccu.callbackNewDevices([]interface{}{
		"go-rf",
		[]interface{}{
			map[string]interface{}{
				"ADDRESS": "address",
				"TYPE":    "switch",
			},
		},
	})
	ass.Equal([]interface{}{true}, resp)
	ass.Nil(fault)
}
