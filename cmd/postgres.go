//go:build postgres

package cmd

import (
	"context"
	"log/slog"
	"time"

	"github.com/niktheblak/ruuvitag-common/pkg/psql"
	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/postgres"
)

func createPostgresExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	psqlInfo, err := psql.CreateConnString(cfg)
	if err != nil {
		return nil, err
	}
	table := cast.ToString(cfg["table"])
	logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to PostgreSQL", slog.String("conn_str", psql.RemovePassword(psqlInfo)), slog.String("table", table))
	return postgres.New(ctx, name, postgres.Config{
		ConnString: psqlInfo,
		Table:      table,
		Columns:    columns,
		Logger:     logger,
	})
}
