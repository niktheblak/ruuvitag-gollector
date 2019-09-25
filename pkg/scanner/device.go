package scanner

import (
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
)

type deviceCreator interface {
	NewDevice() (gatt.Device, error)
}

type gattDeviceCreator struct {
}

func (c gattDeviceCreator) NewDevice() (gatt.Device, error) {
	return gatt.NewDevice(option.DefaultClientOptions...)
}

type peripheralDiscoverer interface {
	HandlePeripheralDiscovered(gatt.Device, func(p gatt.Peripheral, a *gatt.Advertisement, rssi int))
}

type gattPeripheralDiscoverer struct {
}

func (g gattPeripheralDiscoverer) HandlePeripheralDiscovered(d gatt.Device, f func(p gatt.Peripheral, a *gatt.Advertisement, rssi int)) {
	d.Handle(gatt.PeripheralDiscovered(f))
}
