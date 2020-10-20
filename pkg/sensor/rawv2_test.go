package sensor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testData = []byte{
	0x99, 0x04, // Manufacturer ID
	0x05, 0x12, 0xD4, 0x9C, 0x40, 0xC3, 0x40, 0x00, 0x38, 0x00, 0xE4, 0x03, 0xE4, 0x90, 0x76, 0x41,
	0xAD, 0xEE, 0xF7, 0xFA, 0x74, 0x4A, 0x1E, 0x1A, 0xB8,
}

func TestRAWv2Format(t *testing.T) {
	assert.True(t, IsRuuviTag(testData))
}

func TestParseRAWv2Data(t *testing.T) {
	data, err := Parse(testData)
	require.NoError(t, err)
	assert.Equal(t, 24.1, data.Temperature)
	assert.Equal(t, 100.0, data.Humidity)
	assert.Equal(t, 999.84, data.Pressure)
	assert.Equal(t, 2.755, data.BatteryVoltage)
	assert.Equal(t, -18, data.TxPower)
	assert.Equal(t, 56, data.AccelerationX)
	assert.Equal(t, 228, data.AccelerationY)
	assert.Equal(t, 996, data.AccelerationZ)
	assert.Equal(t, 65, data.MovementCounter)
	assert.Equal(t, 44526, data.MeasurementNumber)
}
