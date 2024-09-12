package downsample

import (
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

func Avg(measurements []sensor.Data) sensor.Data {
	if len(measurements) == 0 {
		return sensor.Data{}
	}
	var avgTemperature, avgHumidity, avgPressure, avgDewPoint float64
	for _, measurement := range measurements {
		avgTemperature += measurement.Temperature
		avgHumidity += measurement.Humidity
		avgPressure += measurement.Pressure
		avgDewPoint += measurement.DewPoint
	}
	avgTemperature /= float64(len(measurements))
	avgHumidity /= float64(len(measurements))
	avgPressure /= float64(len(measurements))
	avgDewPoint /= float64(len(measurements))
	m := measurements[len(measurements)-1]
	m.Temperature = avgTemperature
	m.Humidity = avgHumidity
	m.Pressure = avgPressure
	m.DewPoint = avgDewPoint
	return m
}
