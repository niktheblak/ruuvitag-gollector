package ruuvitag

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

type SensorData struct {
	DeviceID      string    `json:"device_id"`
	Name          string    `json:"name"`
	Temperature   float64   `json:"temperature"`
	Humidity      float64   `json:"humidity"`
	Pressure      int       `json:"pressure"`
	Battery       int       `json:"battery"`
	Address       string    `json:"address"`
	AccelerationX int       `json:"acceleration_x"`
	AccelerationY int       `json:"acceleration_y"`
	AccelerationZ int       `json:"acceleration_z"`
	Timestamp     time.Time `json:"ts"`
}

type sensorFormat3 struct {
	ManufacturerID      uint16
	DataFormat          uint8
	Humidity            uint8
	Temperature         uint8
	TemperatureFraction uint8
	Pressure            uint16
	AccelerationX       int16
	AccelerationY       int16
	AccelerationZ       int16
	BatteryVoltageMv    uint16
}

func parseTemperature(t uint8, f uint8) float64 {
	var mask uint8
	mask = 1 << 7
	isNegative := (t & mask) > 0
	temp := float64(t&^mask) + float64(f)/100.0
	if isNegative {
		temp *= -1
	}
	return temp
}

func parseSensorFormat3(data []byte) SensorData {
	reader := bytes.NewReader(data)
	var result sensorFormat3
	err := binary.Read(reader, binary.BigEndian, &result)
	if err != nil {
		panic(err)
	}
	var sd SensorData
	sd.Temperature = parseTemperature(result.Temperature, result.TemperatureFraction)
	sd.Humidity = float64(result.Humidity) / 2.0
	sd.Pressure = int(result.Pressure) + 50000
	sd.Battery = int(result.BatteryVoltageMv)
	sd.AccelerationX = int(result.AccelerationX)
	sd.AccelerationY = int(result.AccelerationY)
	sd.AccelerationZ = int(result.AccelerationZ)
	return sd
}

func Parse(data []byte) (sensorData SensorData, err error) {
	if len(data) == 20 && binary.LittleEndian.Uint16(data[0:2]) == 0x0499 {
		sensorFormat := data[2]
		switch sensorFormat {
		case 3:
			sensorData = parseSensorFormat3(data)
			return
		default:
			err = fmt.Errorf("unknown sensor format: %v", sensorFormat)
			return
		}
	} else {
		err = fmt.Errorf("not a RuuviTag device")
		return
	}
}
