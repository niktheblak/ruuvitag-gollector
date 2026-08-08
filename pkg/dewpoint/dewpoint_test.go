package dewpoint

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
)

func TestCalculate(t *testing.T) {
	dp, err := Calculate(20, temperature.Celsius, 50)
	require.NoError(t, err)
	assert.InDelta(t, 9.3, dp, 0.1)

	dp, err = Calculate(30, temperature.Celsius, 60)
	require.NoError(t, err)
	assert.InDelta(t, 21.4, dp, 0.1)

	_, err = Calculate(-600, temperature.Celsius, 50)
	assert.ErrorIs(t, err, ErrInvalidTemperature)

	_, err = Calculate(400, temperature.Celsius, 50)
	assert.ErrorIs(t, err, ErrInvalidTemperature)

	_, err = Calculate(math.Inf(1), temperature.Kelvin, 50)
	assert.ErrorIs(t, err, ErrInvalidTemperature)

	_, err = Calculate(20, temperature.Kelvin, 0)
	assert.ErrorIs(t, err, ErrInvalidHumidity)

	_, err = Calculate(20, temperature.Kelvin, 101)
	assert.ErrorIs(t, err, ErrInvalidHumidity)
}
