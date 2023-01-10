//go:build !aws

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func addDynamoDBExporter(exporters *[]exporter.Exporter) error {
	return ErrNotEnabled
}

func addSQSExporter(exporters *[]exporter.Exporter) error {
	return ErrNotEnabled
}
