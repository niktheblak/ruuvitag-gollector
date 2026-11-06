package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// exporterInitTimeout is the maximum time an exporter is allowed to take to connect to its backend during startup.
const exporterInitTimeout = 10 * time.Second

// configDuration parses a duration config value.
//
// Plain numbers are interpreted as seconds, while strings are parsed as Go duration
// strings such as "30s" or "5m". A string containing only a number is interpreted as seconds as well.
func configDuration(v any) (time.Duration, error) {
	const unit = time.Second
	switch value := v.(type) {
	case nil:
		return 0, nil
	case time.Duration:
		return value, nil
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return 0, nil
		}
		if n, err := strconv.ParseFloat(value, 64); err == nil {
			return time.Duration(n * float64(unit)), nil
		}
		d, err := time.ParseDuration(value)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", value, err)
		}
		return d, nil
	default:
		n, err := cast.ToFloat64E(v)
		if err != nil {
			return 0, fmt.Errorf("invalid duration %v: %w", v, err)
		}
		return time.Duration(n * float64(unit)), nil
	}
}
