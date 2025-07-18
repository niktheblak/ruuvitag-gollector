package wetbulb

import (
	"testing"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculate(t *testing.T) {
	wb, err := Calculate(29.0, temperature.Celsius, 85.0)
	require.NoError(t, err)
	assert.InDelta(t, 26.9, wb, 0.1)

	_, err = Calculate(29.0, temperature.Celsius, -1)
	assert.ErrorIs(t, err, ErrInvalidHumidity)

	_, err = Calculate(29.0, temperature.Celsius, 150)
	assert.ErrorIs(t, err, ErrInvalidHumidity)
}
