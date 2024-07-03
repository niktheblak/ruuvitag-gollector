//go:build !influxdb

package influxdb

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(cfg Config, logger *slog.Logger) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "InfluxDB"}, nil
}
