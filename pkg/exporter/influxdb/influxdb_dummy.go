//go:build !influxdb

package influxdb

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(cfg Config) exporter.Exporter {
	return exporter.NoOp{ReportedName: "InfluxDB"}
}
