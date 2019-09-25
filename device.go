package homematic

import (
	"sync"

	"github.com/spf13/cast"

	"gitlab.com/bboehmke/homematic/rpc"
)

var devNameScript = `string s_device;
string s_channel;
string output = "";
foreach(s_device, dom.GetObject(ID_DEVICES).EnumIDs()) {
	var o_device = dom.GetObject(s_device);
	output = output # o_device.Address() # "=" # o_device.Name() # "\n" ;
	foreach(s_channel, o_device.Channels().EnumIDs()) {
		var o_channel = dom.GetObject(s_channel);
		output = output # o_channel.Address() # "=" # o_channel.Name() # "\n" ;
	}
}`

// loadDevice from received data
func loadDevice(data map[string]interface{}) *Device {
	flags := cast.ToInt32(data["FLAGS"])
	return &Device{
		Type:      cast.ToString(data["TYPE"]),
		Address:   cast.ToString(data["ADDRESS"]),
		Children:  cast.ToStringSlice(data["CHILDREN"]),
		Parent:    cast.ToString(data["PARENT"]),
		ParamSets: cast.ToStringSlice(data["PARAMSETS"]),

		FlagVisible:    (flags & 0x01) != 0,
		FlagInternal:   (flags & 0x02) != 0,
		FlagDontdelete: (flags & 0x04) != 0,
	}
}

// Device of CCU
type Device struct {
	client            rpc.Client
	valuesDescription map[string]ParameterDescription
	mutex             sync.RWMutex

	Name    string
	Type    string
	Address string

	Children  []string
	Parent    string
	ParamSets []string

	FlagVisible    bool
	FlagInternal   bool
	FlagDontdelete bool

	onValueChange func(key string, value interface{})
}

// nameChanged updates device name
func (d *Device) nameChanged(name string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.Name = name
}

// valueChanged calls OnValueChange function if set
func (d *Device) valueChanged(key string, value interface{}) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if d.onValueChange != nil {
		d.onValueChange(key, value)
	}
}

// nameChanged updates device name
func (d *Device) SetValueChangedHandler(handler func(key string, value interface{})) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.onValueChange = handler
}

// HasValues returns true if device has values
func (d *Device) HasValues() bool {
	for _, p := range d.ParamSets {
		if p == "VALUES" {
			return true
		}
	}
	return false
}

// GetValues of a device
func (d *Device) GetValues() (map[string]interface{}, error) {
	response, err := d.client.Call(
		"getParamset",
		[]interface{}{d.Address, "VALUES"})
	if err != nil {
		return nil, err
	}
	return cast.ToStringMap(response.FirstParam()), nil
}

// GetValue of a device with the given name
func (d *Device) GetValue(name string) (interface{}, error) {
	response, err := d.client.Call(
		"getValue",
		[]interface{}{d.Address, name})
	if err != nil {
		return nil, err
	}
	return response.FirstParam(), nil
}

// SetValue of a device with given name
func (d *Device) SetValue(name string, value interface{}) error {
	_, err := d.client.Call(
		"setValue",
		[]interface{}{d.Address, name, value})
	return err
}

// GetValuesDescription for this device
func (d *Device) GetValuesDescription() (map[string]ParameterDescription, error) {
	// load on first call
	if d.valuesDescription == nil {
		response, err := d.client.Call(
			"getParamsetDescription",
			[]interface{}{d.Address, "VALUES"})
		if err != nil {
			return nil, err
		}

		rawData := cast.ToStringMap(response.FirstParam())
		d.valuesDescription = make(map[string]ParameterDescription, len(rawData))
		for key, value := range rawData {
			d.valuesDescription[key] = loadParameterDescription(value)
		}
	}
	return d.valuesDescription, nil
}

// ParameterDescription contains information about a parameter
type ParameterDescription struct {
	ID       string
	Default  interface{}
	Type     string
	Unit     string
	TabOrder int

	OperationRead  bool
	OperationWrite bool
	OperationEvent bool

	FlagVisible   bool
	FlagInternal  bool
	FlagTransform bool
	FlagService   bool
	FlagSticky    bool
}

// loadParameterDescription from received data
func loadParameterDescription(data interface{}) ParameterDescription {
	dataMap := cast.ToStringMap(data)
	operations := cast.ToInt32(dataMap["OPERATIONS"])
	flags := cast.ToInt32(dataMap["FLAGS"])
	return ParameterDescription{
		ID:       cast.ToString(dataMap["ID"]),
		Default:  dataMap["DEFAULT"],
		Type:     cast.ToString(dataMap["TYPE"]),
		Unit:     cast.ToString(dataMap["UNIT"]),
		TabOrder: cast.ToInt(dataMap["TAB_ORDER"]),

		OperationRead:  (operations & 0x01) != 0,
		OperationWrite: (operations & 0x02) != 0,
		OperationEvent: (operations & 0x04) != 0,

		FlagVisible:   (flags & 0x01) != 0,
		FlagInternal:  (flags & 0x02) != 0,
		FlagTransform: (flags & 0x04) != 0,
		FlagService:   (flags & 0x08) != 0,
		FlagSticky:    (flags & 0x10) != 0,
	}
}
