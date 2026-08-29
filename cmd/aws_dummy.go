//go:build !aws

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createDynamoDBExporter(_ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}

func createSQSExporter(_ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
