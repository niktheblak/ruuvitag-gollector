package wetbulb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculate(t *testing.T) {
	wb, err := Calculate(29.0, 85.0)
	require.NoError(t, err)
	assert.InDelta(t, 26.9, wb, 0.1)

	_, err = Calculate(29.0, -1)
	assert.ErrorIs(t, err, ErrInvalidHumidity)

	_, err = Calculate(29.0, 150)
	assert.ErrorIs(t, err, ErrInvalidHumidity)
}
