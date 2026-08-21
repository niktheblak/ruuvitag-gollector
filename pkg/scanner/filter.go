package scanner

import "github.com/niktheblak/ruuvitag-gollector/pkg/sensor"

func Filter(peripherals map[string]string) AdvertisementFilter {
	return func(a Advertisement) bool {
		if !sensor.IsRuuviTag(a.RawManufacturerData(sensor.RuuviManufacturerID)) {
			return false
		}
		if len(peripherals) == 0 {
			return true
		}
		_, ok := peripherals[NormalizeAddress(a.Address)]
		return ok
	}
}
