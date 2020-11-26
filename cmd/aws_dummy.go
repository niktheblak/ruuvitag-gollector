// +build !aws

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func addDynamoDBExporter(exporters *[]exporter.Exporter) error {
	return nil
}

func addSQSExporter(exporters *[]exporter.Exporter) error {
	return nil
}
