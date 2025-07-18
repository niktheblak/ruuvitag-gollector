package wetbulb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculate(t *testing.T) {
	wb := Calculate(29.0, 85.0)
	assert.InDelta(t, 26.9, wb, 0.1)
}
