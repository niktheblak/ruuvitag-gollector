//go:build !influxdb

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createInfluxDBExporter(columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
