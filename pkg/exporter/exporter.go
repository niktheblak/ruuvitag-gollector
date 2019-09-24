package exporter

import (
	"context"

	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type Exporter interface {
	Name() string
	Export(ctx context.Context, data ruuvitag.SensorData) error
	Close() error
}
