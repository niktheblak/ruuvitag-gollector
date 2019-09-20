package reporter

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type Reporter interface {
	Name() string
	Report(data ruuvitag.SensorData) error
	Close() error
}
