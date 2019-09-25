package console

import (
	"context"
	"fmt"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Exporter struct {
}

func (e Exporter) Name() string {
	return "Console"
}

func (e Exporter) Export(ctx context.Context, data sensor.Data) error {
	fmt.Println(data)
	return nil
}

func (e Exporter) Close() error {
	return nil
}
