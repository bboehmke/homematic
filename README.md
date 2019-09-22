homematic
====
[![GoDoc](https://godoc.org/gitlab.com/bboehmke/homematic?status.svg)](https://godoc.org/gitlab.com/bboehmke/homematic)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/bboehmke/homematic)](https://goreportcard.com/report/gitlab.com/bboehmke/homematic)

homematic is a simple library to interface a [HomeMatic](https://www.homematic.com/) 
CCU2 or CCU3.

The communication is done with XML RPC and supports Wired, RF and HmIP devices.

## Usage

```go
// create client object for CCU
ccu, err := homematic.NewCCU("192.168.4.40")

// list all devices
devices, err := client.GetDevices()

// set state of device
devices["OEQ1234567:1"].SetValue("STATE", true)
````

See the [documentation](https://godoc.org/gitlab.com/bboehmke/homematic) for more information.

