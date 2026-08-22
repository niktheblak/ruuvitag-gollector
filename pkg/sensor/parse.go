package sensor

import (
	"encoding/binary"
	"fmt"

	commonsensor "github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

// RuuviManufacturerID is Ruuvi Innovations' Bluetooth company identifier.
const RuuviManufacturerID uint16 = 0x0499

const (
	DataFormat3ID uint8 = 3
	DataFormat5ID uint8 = 5
)

func Parse(data []byte) (sensorData commonsensor.Data, err error) {
	if !IsRuuviTag(data) {
		err = fmt.Errorf("not a RuuviTag device")
		return
	}
	sensorFormat := data[2]
	switch sensorFormat {
	case DataFormat3ID:
		sensorData, err = ParseSensorFormat3(data)
		return
	case DataFormat5ID:
		sensorData, err = ParseSensorFormat5(data)
		return
	default:
		err = fmt.Errorf("unknown sensor format: %v", sensorFormat)
		return
	}
}

func IsRuuviTag(data []byte) bool {
	return len(data) >= 16 && binary.BigEndian.Uint16(data[0:2]) == 0x9904
}
