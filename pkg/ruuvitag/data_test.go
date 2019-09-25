package ruuvitag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTemperature(t *testing.T) {
	assert.Equal(t, 0.0, ParseTemperature(128, 0), "Negative zero is zero")
	assert.Equal(t, -2.0, ParseTemperature(130, 0), "-2")
	assert.Equal(t, 2.0, ParseTemperature(2, 0), "2")
	assert.Equal(t, 2.2, ParseTemperature(2, 20), "fraction is ok")
	assert.Equal(t, -2.99, ParseTemperature(130, 99), "fraction is ok negative")
}
