//go:build !gcp

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createPubSubExporter(columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
