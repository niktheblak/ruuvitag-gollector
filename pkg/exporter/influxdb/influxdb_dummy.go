//go:build !influxdb

package influxdb

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(_ Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "InfluxDB"}, nil
}
