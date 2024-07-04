//go:build !gcp

package pubsub

import (
	"context"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "Google Pub/Sub"}, nil
}
