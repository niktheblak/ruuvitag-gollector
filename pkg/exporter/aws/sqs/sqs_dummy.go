// +build !aws

package sqs

import (
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(cfg Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "AWS SQS"}, nil
}
