package influxdb

import (
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
}
