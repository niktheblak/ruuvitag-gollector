//go:build !postgres

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createPostgresExporter(_ string, _ map[string]string, _ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
