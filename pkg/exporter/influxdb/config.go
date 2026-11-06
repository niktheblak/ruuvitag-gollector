package influxdb

import (
	"fmt"
	"log/slog"
	"time"
)

const DefaultMeasurement = "ruuvitag"

type Config struct {
	Addr          string
	Org           string
	Bucket        string
	Measurement   string
	Token         string
	Async         bool
	BatchSize     int
	FlushInterval time.Duration
	Columns       map[string]string
	Logger        *slog.Logger
}

func Validate(cfg Config) error {
	if cfg.Addr == "" {
		return fmt.Errorf("InfluxDB address must be specified")
	}
	if cfg.Org == "" {
		return fmt.Errorf("InfluxDB organization must be specified")
	}
	if cfg.Bucket == "" {
		return fmt.Errorf("InfluxDB bucket must be specified")
	}
	if cfg.Token == "" {
		return fmt.Errorf("InfluxDB token must be specified")
	}
	if len(cfg.Columns) == 0 {
		return fmt.Errorf("columns must be non-empty")
	}
	return nil
}
