package scanner

import (
	"tinygo.org/x/bluetooth"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

func Filter(peripherals map[string]string, result bluetooth.ScanResult) bool {
	md := result.ManufacturerData()
	if len(md) == 0 {
		return false
	}
	for _, m := range md {
		if !sensor.IsRuuviTag(m.Data) {
			return false
		}
	}
	if peripherals == nil || len(peripherals) == 0 {
		return true
	}
	_, ok := peripherals[result.Address.String()]
	return ok
}
