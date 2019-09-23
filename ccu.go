package homematic

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cast"

	"gitlab.com/bboehmke/homematic/rpc"
	"gitlab.com/bboehmke/homematic/script"
)

// NewCCU creates a new connection to a CCU
func NewCCU(address string) (*CCU, error) {
	ccu := &CCU{
		rpcClients: map[string]rpc.Client{
			"wired": rpc.NewClient(fmt.Sprintf("http://%s:2000/", address)),
			"rf":    rpc.NewClient(fmt.Sprintf("http://%s:2001/", address)),
			"hmIP":  rpc.NewClient(fmt.Sprintf("http://%s:2010/", address)),
		},
		scriptClient: script.NewClient(fmt.Sprintf("http://%s:8181/", address)),
		devices:      make(map[string]*Device),
	}
	ccu.lastEvent = make(map[string]time.Time, len(ccu.rpcClients))

	// prepare RPC server
	var err error
	ccu.rpcServer, err = rpc.NewServer(ccu.handleEvents)
	return ccu, err
}

// CCU represents a connection to a Homematic CCU
type CCU struct {
	rpcClients map[string]rpc.Client
	rpcServer  *rpc.Server

	scriptClient script.Client

	mutex      sync.RWMutex
	devices    map[string]*Device
	lastUpdate time.Time
	lastEvent  map[string]time.Time
}

// handle received events
func (c *CCU) handleEvents(method string, params []interface{}) ([]interface{}, *rpc.Fault) {
	if method != "event" {
		return nil, nil
	}
	if len(params) < 4 {
		return nil, &rpc.Fault{
			Code:   -1,
			String: "invalid event call",
		}
	}

	c.mutex.RLock()
	c.lastEvent[cast.ToString(params[0])] = time.Now()

	// if device is known trigger value change
	device, ok := c.devices[cast.ToString(params[1])]
	c.mutex.RUnlock()
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

// checkEventHandling for activity and re init if no events since long time
func (c *CCU) checkEventHandling() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// check only if server is running
	if !c.rpcServer.IsRunning() {
		return nil
	}

	for id, client := range c.rpcClients {
		// only re init if no event since 10 minutes
		if time.Since(c.lastEvent[id]) < time.Minute*10 {
			continue
		}

		ip, err := client.LocalIP()
		if err != nil {
			return err
		}

		response, err := client.Call("init", []interface{}{
			fmt.Sprintf("http://%s:%d", ip, c.rpcServer.Port()),
			id,
		})
		if err != nil {
			return err
		}

		if response.Fault != nil {
			return errors.New(response.Fault.String)
		}
		c.lastEvent[id] = time.Now()
	}
	return nil
}

// Start event handling
func (c *CCU) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.rpcServer.Start()

	for id, client := range c.rpcClients {
		ip, err := client.LocalIP()
		if err != nil {
			return err
		}

		// ignore result -> handle all clients
		_, _ = client.Call("init", []interface{}{
			fmt.Sprintf("http://%s:%d", ip, c.rpcServer.Port()),
			id,
		})
		c.lastEvent[id] = time.Now()
	}
	return nil
}

// Stop event handling
func (c *CCU) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, client := range c.rpcClients {
		ip, err := client.LocalIP()
		if err != nil {
			return err
		}

		_, _ = client.Call("init", []interface{}{
			fmt.Sprintf("http://%s:%d", ip, c.rpcServer.Port()),
			"",
		})
	}
	return c.rpcServer.Stop()
}

// GetDevices from CCU
func (c *CCU) GetDevices() (map[string]*Device, error) {
	err := c.UpdateDevices(false)
	if err != nil {
		return nil, err
	}

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.devices, nil
}

// UpdateDevices currently known on CCU
func (c *CCU) UpdateDevices(force bool) error {
	err := c.checkEventHandling()
	if err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// update only every 10 minutes or if force is set
	if !force && time.Since(c.lastUpdate) < time.Minute*10 {
		return nil
	}

	// get device names from logic layer
	scriptData, err := c.scriptClient.Call(devNameScript)
	if err != nil {
		return err
	}
	deviceNames := scriptData.GetMap("output")

	c.lastUpdate = time.Now()
	currentDevices := make(map[string]bool, len(deviceNames))
	// iterate over all interfaces
	for _, client := range c.rpcClients {
		response, err := client.Call("listDevices", nil)
		if err != nil {
			return err
		}

		// load each device
		for _, data := range cast.ToSlice(response.FirstParam()) {
			device := loadDevice(cast.ToStringMap(data))
			_, ok := c.devices[device.Address]
			if !ok {
				device.client = client
				c.devices[device.Address] = device
			}
			currentDevices[device.Address] = true

			c.devices[device.Address].nameChanged(deviceNames[device.Address])
		}
	}
	// cleanup
	for address := range c.devices {
		if !currentDevices[address] {
			delete(c.devices, address)
		}
	}

	return nil
}
