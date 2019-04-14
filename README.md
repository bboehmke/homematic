homematic
====
[![GoDoc](https://godoc.org/gitlab.com/bboehmke/homematic?status.svg)](https://godoc.org/gitlab.com/bboehmke/homematic)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/bboehmke/homematic)](https://goreportcard.com/report/gitlab.com/bboehmke/homematic)

homematic is a simple library to interface a [HomeMatic](https://www.homematic.com/) 
CCU2 or CCU3.

The communication is done with XML RPC and supports Wired and wireless devices.

## Usage

```go
// create client object for CCU
client := homematic.NewClient("192.168.4.40")

// list all wireless devices
device, err := client.RF.ListDevices()

// set state of wired device
err := client.Wired.SetValue("OEQ1234567:1", "STATE", true)
````

See the [documentation](https://godoc.org/gitlab.com/bboehmke/homematic) for more information.

