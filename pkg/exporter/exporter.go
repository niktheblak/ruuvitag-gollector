package exporter

import (
	"context"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Exporter interface {
	Name() string
	Export(ctx context.Context, data sensor.Data) error
	Close() error
}
