//go:build !gcp

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createPubSubExporter(_ string, _ map[string]string, _ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
