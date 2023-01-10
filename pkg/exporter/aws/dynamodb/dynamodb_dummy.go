//go:build !aws

package dynamodb

import "github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

func New(cfg Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "AWS DynamoDB"}, nil
}
