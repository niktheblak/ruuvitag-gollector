//go:build !influxdb

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func addInfluxDBExporter(exporters *[]exporter.Exporter) error {
	return ErrNotEnabled
}
