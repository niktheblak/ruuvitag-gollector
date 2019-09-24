package config

import (
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExampleConfig(t *testing.T) {
	var config Config
	blob := `
		reporting_interval = "1m"

		[[ruuvitag]]
		mac = "CC:CA:7E:52:CC:34"
		name = "Backyard"
		
		[[ruuvitag]]
		mac = "FB:E1:B7:04:95:EE"
		name = "Upstairs"
	`
	_, err := toml.Decode(blob, &config)
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, config.ReportingInterval.Duration)
	require.Len(t, config.RuuviTag, 2)
	assert.Equal(t, "CC:CA:7E:52:CC:34", config.RuuviTag[0].MAC)
	assert.Equal(t, "Backyard", config.RuuviTag[0].Name)
	assert.Equal(t, "FB:E1:B7:04:95:EE", config.RuuviTag[1].MAC)
	assert.Equal(t, "Upstairs", config.RuuviTag[1].Name)
}

func TestValidateMissingReportingInterval(t *testing.T) {
	cfg := Config{
		RuuviTag: []RuuviTag{
			{
				MAC:  "CC:CA:7E:52:CC:34",
				Name: "Backyard",
			},
		},
	}
	err := cfg.Validate()
	assert.Error(t, err)
}

func TestValidate(t *testing.T) {
	cfg := Config{
		ReportingInterval: Duration{60 * time.Second},
		RuuviTag: []RuuviTag{
			{
				MAC:  "CC:CA:7E:52:CC:34",
				Name: "Backyard",
			},
		},
	}
	err := cfg.Validate()
	assert.NoError(t, err)
}
