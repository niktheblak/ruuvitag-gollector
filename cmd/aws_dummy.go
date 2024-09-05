//go:build !aws

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createDynamoDBExporter(cfg map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}

func createSQSExporter(cfg map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
