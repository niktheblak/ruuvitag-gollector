//go:build !gcp

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func initStackdriverLogging() error {
	return ErrNotEnabled
}

func addPubSubExporter(exporters *[]exporter.Exporter, columns map[string]string) error {
	return ErrNotEnabled
}
