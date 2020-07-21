package dewpoint

import (
	"testing"

	"github.com/niktheblak/ruuvitag-gollector/pkg/temperature"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculate(t *testing.T) {
	dp, err := Calculate(20, temperature.Celsius, 50)
	require.NoError(t, err)
	assert.InDelta(t, 9.3, dp, 0.1)

	dp, err = Calculate(30, temperature.Celsius, 60)
	require.NoError(t, err)
	assert.InDelta(t, 21.4, dp, 0.1)
}
