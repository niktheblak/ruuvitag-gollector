//go:build !timescaledb

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func addTimescaleDBExporter(exporters *[]exporter.Exporter) error {
	return ErrNotEnabled
}
