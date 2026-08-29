//go:build !aws

package dynamodb

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func New(_ Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "AWS DynamoDB"}, nil
}
