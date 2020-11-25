package exporter

import (
	"context"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type NoOp struct {
	ReportedName string
}

func (e NoOp) Name() string {
	return e.ReportedName
}

func (e NoOp) Export(ctx context.Context, data sensor.Data) error {
	return nil
}

func (e NoOp) Close() error {
	return nil
}
