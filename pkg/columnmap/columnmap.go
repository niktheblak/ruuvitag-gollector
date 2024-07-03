package columnmap

import (
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

// Collect calls a function on each sensor data field, providing the preferred column name for the field.
func Collect(columns map[string]string, data sensor.Data, f func(column string, v any)) {
	for _, c := range sensor.DefaultColumns {
		cn, ok := columns[c]
		if !ok {
			continue
		}
		switch c {
		case "time":
			f(cn, data.Timestamp)
		case "mac":
			f(cn, data.Addr)
		case "name":
			f(cn, data.Name)
		case "temperature":
			f(cn, data.Temperature)
		case "humidity":
			f(cn, data.Humidity)
		case "pressure":
			f(cn, data.Pressure)
		case "acceleration_x":
			f(cn, data.AccelerationX)
		case "acceleration_y":
			f(cn, data.AccelerationY)
		case "acceleration_z":
			f(cn, data.AccelerationZ)
		case "movement_counter":
			f(cn, data.MovementCounter)
		case "measurement_number":
			f(cn, data.MeasurementNumber)
		case "dew_point":
			f(cn, data.DewPoint)
		case "battery_voltage":
			f(cn, data.BatteryVoltage)
		case "tx_power":
			f(cn, data.TxPower)
		}
	}
}
