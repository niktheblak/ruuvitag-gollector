//go:build !mqtt

package mqtt

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(cfg Config) exporter.Exporter {
	return exporter.NoOp{ReportedName: "MQTT"}
}
