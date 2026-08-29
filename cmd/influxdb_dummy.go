//go:build !influxdb

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createInfluxDBExporter(_ map[string]string, _ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
