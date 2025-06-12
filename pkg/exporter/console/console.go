package console

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type consoleExporter struct {
	name string
}

func New(name string) exporter.Exporter {
	return &consoleExporter{name: name}
}

func (e *consoleExporter) Name() string {
	return e.name
}

func (e *consoleExporter) Export(ctx context.Context, data sensor.Data) error {
	j, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(j))
	return nil
}

func (e *consoleExporter) Close() error {
	return nil
}
