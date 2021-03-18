// +build !mqtt

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func addMQTTExporter(exporters *[]exporter.Exporter) error {
	return ErrNotEnabled
}
