package scanner

import (
	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

func Filter(peripherals map[string]string) func(ble.Advertisement) bool {
	return func(a ble.Advertisement) bool {
		if !sensor.IsRuuviTag(a.ManufacturerData()) {
			return false
		}
		if len(peripherals) == 0 {
			return true
		}
		_, ok := peripherals[a.Addr().String()]
		return ok
	}
}
