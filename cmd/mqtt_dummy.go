//go:build !mqtt

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createMQTTExporter(cfg map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
