package sensor

import (
	"encoding/binary"
	"fmt"
	"time"
)

type Data struct {
	Addr            string    `json:"mac"`
	Name            string    `json:"name"`
	Temperature     float64   `json:"temperature"`
	Humidity        float64   `json:"humidity"`
	Pressure        float64   `json:"pressure"`
	Battery         int       `json:"battery"`
	AccelerationX   int       `json:"acceleration_x"`
	AccelerationY   int       `json:"acceleration_y"`
	AccelerationZ   int       `json:"acceleration_z"`
	MovementCounter int       `json:"movement_counter"`
	Timestamp       time.Time `json:"ts"`
}

func Parse(data []byte) (sensorData Data, err error) {
	if !IsRuuviTag(data) {
		err = fmt.Errorf("not a RuuviTag device")
		return
	}
	sensorFormat := data[2]
	switch sensorFormat {
	case 3:
		sensorData, err = ParseSensorFormat3(data)
		return
	case 5:
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
