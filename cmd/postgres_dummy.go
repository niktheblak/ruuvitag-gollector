//go:build !postgres

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createPostgresExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
