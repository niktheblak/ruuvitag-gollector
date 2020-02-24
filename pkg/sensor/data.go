package sensor

import (
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
