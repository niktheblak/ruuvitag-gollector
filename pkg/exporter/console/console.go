package console

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type Exporter struct {
}

func (e Exporter) Name() string {
	return "Console"
}

func (e Exporter) Export(ctx context.Context, data sensor.Data) error {
	j, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(string(j))
	return nil
}

func (e Exporter) Close() error {
	return nil
}
