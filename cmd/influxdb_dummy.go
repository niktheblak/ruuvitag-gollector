//go:build !influxdb

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func addInfluxDBExporter(exporters *[]exporter.Exporter, columns map[string]string) error {
	return ErrNotEnabled
}
