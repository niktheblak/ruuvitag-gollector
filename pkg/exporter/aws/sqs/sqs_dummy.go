//go:build !aws

package sqs

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(_ Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "AWS SQS"}, nil
}
