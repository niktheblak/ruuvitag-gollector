//go:build !postgres

package postgres

import (
	"context"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "Postgres"}, nil
}
