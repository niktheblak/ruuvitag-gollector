// +build !gcp

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func initStackdriverLogging() error {
	return nil
}

func addPubSubExporter(exporters *[]exporter.Exporter) error {
	return nil
}
