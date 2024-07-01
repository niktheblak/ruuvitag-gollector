package scanner

import (
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

type DeviceCreator interface {
	NewDevice(impl string) (ble.Device, error)
}

type GoBLEDeviceCreator struct {
}

func (c *GoBLEDeviceCreator) NewDevice(impl string) (ble.Device, error) {
	d, err := dev.NewDevice(impl)
	if err != nil {
		return nil, err
	}
	ble.SetDefaultDevice(d)
	return d, nil
}
