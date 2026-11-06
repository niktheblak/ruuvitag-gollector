//go:build !aws

package cmd

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func createDynamoDBExporter(_ string, _ map[string]string, _ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}

func createSQSExporter(_ string, _ map[string]string, _ map[string]any) (exporter.Exporter, error) {
	return nil, ErrNotEnabled
}
