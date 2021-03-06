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
	return NewCCUCustom(address, "go")
}

// NewCCUCustom creates a new connection to a CCU with custom id
func NewCCUCustom(address, id string) (*CCU, error) {
	ccu := &CCU{
		rpcClients: map[string]rpc.Client{
			fmt.Sprintf("%s-wired", id): rpc.NewClient(fmt.Sprintf("http://%s:2000/", address)),
			fmt.Sprintf("%s-rf", id):    rpc.NewClient(fmt.Sprintf("http://%s:2001/", address)),
			fmt.Sprintf("%s-hmip", id):  rpc.NewClient(fmt.Sprintf("http://%s:2010/", address)),
		},
		scriptClient: script.NewClient(fmt.Sprintf("http://%s:8181/", address)),
		devices:      make(map[string]*Device),
	}
	ccu.lastClientEvent = make(map[string]time.Time, len(ccu.rpcClients))

	// prepare RPC server
	var err error
	ccu.rpcServer, err = rpc.NewServer(ccu.handleCallback)
	return ccu, err
}

// CCU represents a connection to a Homematic CCU
type CCU struct {
	rpcClients      map[string]rpc.Client
	rpcServer       *rpc.Server
	lastClientEvent map[string]time.Time

	scriptClient script.Client
	clientMutex  sync.RWMutex

	devices     map[string]*Device
	lastUpdate  time.Time
	deviceMutex sync.RWMutex
}

// checkEventHandling for activity and re init if no events since long time
func (c *CCU) checkEventHandling() error {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

	// check only if server is running
	if !c.rpcServer.IsRunning() {
		return nil
	}

	for id, client := range c.rpcClients {
		// only re init if no event since 10 minutes
		if time.Since(c.lastClientEvent[id]) < time.Minute*10 {
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
		c.lastClientEvent[id] = time.Now()
	}
	return nil
}

// Start event handling
func (c *CCU) Start() error {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

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
		c.lastClientEvent[id] = time.Now()
	}
	return nil
}

// Stop event handling
func (c *CCU) Stop() error {
	c.clientMutex.Lock()
	defer c.clientMutex.Unlock()

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

	c.deviceMutex.RLock()
	defer c.deviceMutex.RUnlock()

	return c.devices, nil
}

// UpdateDevices currently known on CCU
func (c *CCU) UpdateDevices(force bool) error {
	err := c.checkEventHandling()
	if err != nil {
		return err
	}

	c.deviceMutex.Lock()
	defer c.deviceMutex.Unlock()

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
