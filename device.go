package homematic

import "github.com/spf13/cast"

var devNameScript = `string s_device;
string s_channel;
string output = "";
foreach(s_device, dom.GetObject(ID_DEVICES).EnumUsedIDs()) {
	var o_device = dom.GetObject(s_device);
	output = output # o_device.Address() # "=" # o_device.Name() # "\n" ;
	foreach(s_channel, o_device.Channels().EnumUsedIDs()) {
		var o_channel = dom.GetObject(s_channel);
		output = output # o_channel.Address() # "=" # o_channel.Name() # "\n" ;
	}
}`

// ListDevices that are available
func (c *BaseClient) ListDevices() ([]DeviceDescription, error) {
	scriptData, err := c.script.Call(devNameScript)
	if err != nil {
		return nil, err
	}
	deviceNames := scriptData.GetMap("output")

	response, err := c.rpc.Call(
		"listDevices",
		nil)
	if err != nil {
		return nil, err
	}

	rawData := cast.ToSlice(response.FirstParam())
	devices := make([]DeviceDescription, len(rawData))
	for i, v := range rawData {
		devices[i] = loadDeviceDescription(v, deviceNames)
	}
	return devices, nil
}

// DeviceDescription contains information about device
type DeviceDescription struct {
	Name    string
	Type    string
	Address string

	Children  []string
	Parent    string
	ParamSets []string

	/*
		RFAddress int
		ParentType string
		Index int
		AESActive bool
		Firmware string
		AvailableFirmware string
		Updatable bool
		Version int
		Flags int
		LinkSourceRoles string
		LinkTargetRoles string
		Direction int
		Group string
		Team string
		TeamTag string
		TeamChannel []string
		Interface string
		Roaming bool
		RXMode int
	*/
}

// loadDeviceDescription creates DeviceDescription from received data
func loadDeviceDescription(data interface{}, deviceNames map[string]string) DeviceDescription {
	device := DeviceDescription{}
	for key, value := range cast.ToStringMap(data) {
		switch key {
		case "TYPE":
			device.Type = cast.ToString(value)
		case "ADDRESS":
			device.Address = cast.ToString(value)
			device.Name = deviceNames[device.Address]

		case "CHILDREN":
			device.Children = cast.ToStringSlice(value)
		case "PARENT":
			device.Parent = cast.ToString(value)

		case "PARAMSETS":
			device.ParamSets = cast.ToStringSlice(value)
		}
	}
	return device
}

// GetValues returns all values of a device
func (c *BaseClient) GetValues(address string) (map[string]interface{}, error) {
	response, err := c.rpc.Call(
		"getParamset",
		[]interface{}{address, "VALUES"})
	if err != nil {
		return nil, err
	}
	return cast.ToStringMap(response.FirstParam()), nil
}

// GetValue returns a specific value of a device
func (c *BaseClient) GetValue(address, state string) (interface{}, error) {
	response, err := c.rpc.Call(
		"getValue",
		[]interface{}{address, state})
	if err != nil {
		return nil, err
	}
	return response.FirstParam(), nil
}

// SetValue sets a specific value of a device
func (c *BaseClient) SetValue(address, state string, value interface{}) error {
	_, err := c.rpc.Call(
		"setValue",
		[]interface{}{address, state, value})
	return err
}
