//go:build !mqtt

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createMQTTExporter(_ string, _ map[string]string, _ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
