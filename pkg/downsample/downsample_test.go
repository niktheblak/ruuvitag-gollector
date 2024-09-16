package downsample

import (
	"testing"
	"time"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/stretchr/testify/assert"
)

var testMeasurement = sensor.Data{
	Timestamp:         time.Date(2024, time.September, 16, 12, 0, 0, 0, time.UTC),
	Name:              "test-1",
	Addr:              "cc:ca:7e:52:cc:34",
	Temperature:       20,
	Humidity:          60,
	Pressure:          998,
	DewPoint:          13,
	AccelerationX:     948,
	AccelerationY:     408,
	AccelerationZ:     -32,
	MovementCounter:   121,
	MeasurementNumber: 37241,
	BatteryVoltage:    3.006,
	TxPower:           -18,
}

func TestAvg(t *testing.T) {
	m1 := testMeasurement
	m2 := testMeasurement
	m2.Timestamp = m2.Timestamp.Add(1 * time.Minute)
	m2.Temperature = 21
	m2.Humidity = 63
	m2.DewPoint = 13.5
	m2.MeasurementNumber = m2.MeasurementNumber + 1
	m3 := testMeasurement
	m3.Timestamp = m3.Timestamp.Add(2 * time.Minute)
	m3.Temperature = 22
	m3.Humidity = 65
	m3.DewPoint = 14
	m3.MeasurementNumber = m3.MeasurementNumber + 2
	measurements := []sensor.Data{m1, m2, m3}

	avg := Avg(measurements)

	assert.Equal(t, m3.Timestamp, avg.Timestamp)
	assert.Equal(t, testMeasurement.Name, avg.Name)
	assert.Equal(t, testMeasurement.Addr, avg.Addr)
	assert.Equal(t, 21.0, avg.Temperature)
	assert.Equal(t, 62.666666666666664, avg.Humidity)
	assert.Equal(t, 998.0, avg.Pressure)
	assert.Equal(t, 13.5, avg.DewPoint)
	assert.Equal(t, testMeasurement.AccelerationX, avg.AccelerationX)
	assert.Equal(t, testMeasurement.AccelerationY, avg.AccelerationY)
	assert.Equal(t, testMeasurement.AccelerationZ, avg.AccelerationZ)
	assert.Equal(t, testMeasurement.MovementCounter, avg.MovementCounter)
	assert.Equal(t, testMeasurement.MeasurementNumber+2, avg.MeasurementNumber)
	assert.Equal(t, testMeasurement.BatteryVoltage, avg.BatteryVoltage)
	assert.Equal(t, testMeasurement.TxPower, avg.TxPower)
}
