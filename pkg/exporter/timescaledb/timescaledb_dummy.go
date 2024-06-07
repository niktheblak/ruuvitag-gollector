//go:build !timescaledb

package timescaledb

import (
	"context"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

func New(ctx context.Context, connStr, table string) (exporter.Exporter, error) {
	return exporter.NoOp{ReportedName: "Postgres"}, nil
}
