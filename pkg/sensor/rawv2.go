package sensor

import (
	"bytes"
	"encoding/binary"

	"github.com/niktheblak/ruuvitag-gollector/pkg/dewpoint"
	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

/* Payload:
Byte    Value Range			Explanation
---------------------------------------
0 		05 					Format type code
1–2 	-40.000 — 84.995 	Temperature (16bit signed in .005 centigrade. ) -32768 aka 0x8000 Invalid / not available
Examples: 0x4268 = 17000, 17000*.005= 85; 0x1388 = 5000, 5000*.005=25
0xFF67 (two's complement) = 0x0098+1 i.e. 153 , 153*.005 = -0.765

3–4 	0 — 100				Humidity: 16bit unsigned; in .0025%, divide by 400 to get percent.
Example 0x9470 = 38000/400 = 95% . 0xBB80 (i.e. 120) indicates invalid.

5–6 	300 — 11,000 		atmospheric pressure ( 16bit unsigned, value max 50kPa) 65535 (0xFFFF) indicates invalid or unavailable
7–8 	-16,000 — 16,000 	Acceleration-X ( 16bit signed Most Significant Byte first)
9–10 	-16,000 — 16,000 	Acceleration-Y STMicroelectronics LIS2DH12
11–12 	-16,000 — 16,000 	Acceleration-Z

Examples:03F8 = 1.016
( values over 16000 indicates error (implementation pending 10/04/17)

13–14.2	1.6 — 3.646 		Battery voltage above 1.6V, in millivolts, 11 bits unsigned (0-2046)
2047 (FFEx or FFFx) indicates an invalid reading.

14.3–14.7
byte 14&1F	-40 — +20 		TX power above -40dBm, in 2dBm steps. 5 bits unsigned. Value of 31 (0x1F) indicates invalid value.(?)

15		0 — 254 			Movement counter (8bit unsigned), incremented by motion detection interrupts from LIS2DH12 Accelerometer
16–17 	0 — 65,534 			Measurement sequence number (16bit unsigned).
18–23 	00:00:00:...		-
*/
type DataFormat5 struct {
	ManufacturerID    uint16
	DataFormat        uint8
	Temperature       int16
	Humidity          uint16
	Pressure          uint16
	AccelerationX     int16
	AccelerationY     int16
	AccelerationZ     int16
	BatteryVoltage    uint16
	MovementCounter   uint8
	MeasurementNumber uint16
}

func ParseSensorFormat5(data []byte) (sd Data, err error) {
	reader := bytes.NewReader(data)
	var result DataFormat5
	err = binary.Read(reader, binary.BigEndian, &result)
	if err != nil {
		return
	}
	sd.Temperature = float64(result.Temperature) * 0.005
	sd.Humidity = float64(result.Humidity) / 400.0
	sd.DewPoint, _ = dewpoint.Calculate(sd.Temperature, temperature.Celsius, sd.Humidity)
	sd.Pressure = float64(int(result.Pressure)+50000) / 100.0
	sd.AccelerationX = int(result.AccelerationX)
	sd.AccelerationY = int(result.AccelerationY)
	sd.AccelerationZ = int(result.AccelerationZ)
	sd.Battery = int(result.BatteryVoltage >> 5)
	sd.MovementCounter = int(result.MovementCounter)
	return
}
