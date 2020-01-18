package homematic

import (
	"time"

	"github.com/spf13/cast"

	"gitlab.com/bboehmke/homematic/rpc"
)

// handle received callback request
func (c *CCU) handleCallback(method string, params []interface{}) ([]interface{}, *rpc.Fault) {
	switch method {
	case "event":
		return c.callbackEvent(params)
	case "listDevices":
		return c.callbackListDevices()
	case "newDevices":
		return c.callbackNewDevices(params)
	}
	return []interface{}{true}, nil
}

// handle event callback
func (c *CCU) callbackEvent(params []interface{}) ([]interface{}, *rpc.Fault) {
	if len(params) < 4 {
		return nil, &rpc.Fault{
			Code:   -1,
			String: "invalid event call",
		}
	}

	c.mutex.Lock()
	c.lastEvent[cast.ToString(params[0])] = time.Now()

	// if device is known trigger value change
	device, ok := c.devices[cast.ToString(params[1])]
	c.mutex.Unlock()
	if ok {
		device.valueChanged(cast.ToString(params[2]), params[3])

		// check if event handling is working
		c.checkEventHandling()
	} else {
		// if devices does not exist update device list
		c.UpdateDevices(true)
	}
	return nil, nil
}

// handle listDevices callback
func (c *CCU) callbackListDevices() ([]interface{}, *rpc.Fault) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	data := make([]interface{}, 0, len(c.devices))
	for _, device := range c.devices {
		data = append(data, map[string]interface{}{
			"ADDRESS": device.Address,
			"VERSION": device.Version,
		})
	}

	return []interface{}{
		data,
	}, nil
}

// handle newDevices callback
func (c *CCU) callbackNewDevices(params []interface{}) ([]interface{}, *rpc.Fault) {
	if len(params) < 2 {
		return nil, &rpc.Fault{
			Code:   -1,
			String: "invalid newDevices call",
		}
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check if client is known
	client, ok := c.rpcClients[cast.ToString(params[0])]
	if !ok {
		return nil, &rpc.Fault{
			Code:   -1,
			String: "invalid interface id",
		}
	}

	// get device names from logic layer
	scriptData, err := c.scriptClient.Call(devNameScript)
	if err != nil {
		// failed to get device names
		return []interface{}{true}, nil
	}
	deviceNames := scriptData.GetMap("output")

	// load each device
	for _, data := range cast.ToSlice(params[1]) {
		device := loadDevice(cast.ToStringMap(data))
		_, ok := c.devices[device.Address]
		if !ok {
			device.client = client
			c.devices[device.Address] = device
		}

		c.devices[device.Address].nameChanged(deviceNames[device.Address])
	}

	return []interface{}{true}, nil
}
