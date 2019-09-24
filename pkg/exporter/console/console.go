package console

import (
	"context"
	"fmt"

	"github.com/niktheblak/ruuvitag-gollector/pkg/ruuvitag"
)

type Exporter struct {
}

func (e Exporter) Name() string {
	return "Console"
}

func (e Exporter) Export(ctx context.Context, data ruuvitag.SensorData) error {
	fmt.Println(data)
	return nil
}

func (e Exporter) Close() error {
	return nil
}
