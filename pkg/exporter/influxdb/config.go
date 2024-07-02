package influxdb

import (
	"fmt"
	"log/slog"
	"time"
)

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
		return fmt.Errorf("InfluxDB host must be specified")
	}
	if cfg.Bucket == "" {
		return fmt.Errorf("bucket must be specified")
	}
	if cfg.Token == "" {
		return fmt.Errorf("token must be specified")
	}
	return nil
}
