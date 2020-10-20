package sensor

import (
	"time"
)

type Data struct {
	Addr              string    `json:"mac"`
	Name              string    `json:"name"`
	Temperature       float64   `json:"temperature"`
	Humidity          float64   `json:"humidity"`
	DewPoint          float64   `json:"dew_point,omitempty"`
	Pressure          float64   `json:"pressure"`
	BatteryVoltage    float64   `json:"battery_voltage,omitempty"`
	TxPower           int       `json:"tx_power,omitempty"`
	AccelerationX     int       `json:"acceleration_x"`
	AccelerationY     int       `json:"acceleration_y"`
	AccelerationZ     int       `json:"acceleration_z"`
	MovementCounter   int       `json:"movement_counter"`
	MeasurementNumber int       `json:"measurement_number"`
	Timestamp         time.Time `json:"ts"`
}
