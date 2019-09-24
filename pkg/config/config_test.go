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
		id = "CC:CA:7E:52:CC:34"
		name = "Backyard"
		
		[[ruuvitag]]
		id = "FB:E1:B7:04:95:EE"
		name = "Upstairs"
	`
	_, err := toml.Decode(blob, &config)
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, config.ReportingInterval.Duration)
	require.Len(t, config.RuuviTags, 2)
	assert.Equal(t, "CC:CA:7E:52:CC:34", config.RuuviTags[0].ID)
	assert.Equal(t, "Backyard", config.RuuviTags[0].Name)
	assert.Equal(t, "FB:E1:B7:04:95:EE", config.RuuviTags[1].ID)
	assert.Equal(t, "Upstairs", config.RuuviTags[1].Name)
}

func TestValidateMissingReportingInterval(t *testing.T) {
	cfg := Config{
		RuuviTags: []RuuviTag{
			{
				ID:   "CC:CA:7E:52:CC:34",
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
		RuuviTags: []RuuviTag{
			{
				ID:   "CC:CA:7E:52:CC:34",
				Name: "Backyard",
			},
		},
	}
	err := cfg.Validate()
	assert.NoError(t, err)
}
