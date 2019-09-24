package exporter

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type Exporter interface {
	Name() string
	Export(data ruuvitag.SensorData) error
	Close() error
}
