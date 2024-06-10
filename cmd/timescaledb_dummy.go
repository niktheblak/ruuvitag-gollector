//go:build !timescaledb

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func addPostgresExporter(exporters *[]exporter.Exporter) error {
	return ErrNotEnabled
}
