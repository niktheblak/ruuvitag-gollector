package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDuration(t *testing.T) {
	tests := []struct {
		value    any
		expected time.Duration
	}{
		{nil, 0},
		{"", 0},
		{30, 30 * time.Second},
		{int64(30), 30 * time.Second},
		{1.5, 1500 * time.Millisecond},
		{"30", 30 * time.Second},
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"1h30m", 90 * time.Minute},
		{2 * time.Minute, 2 * time.Minute},
	}
	for _, tt := range tests {
		d, err := configDuration(tt.value)
		require.NoError(t, err, "value %v", tt.value)
		assert.Equal(t, tt.expected, d, "value %v", tt.value)
	}
	_, err := configDuration("soon")
	require.Error(t, err)
	_, err = configDuration([]string{"30"})
	require.Error(t, err)
}
