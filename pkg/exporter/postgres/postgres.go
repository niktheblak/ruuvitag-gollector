//go:build postgres

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"

	_ "github.com/lib/pq"
)

type postgresExporter struct {
	db         *sql.DB
	insertStmt *sql.Stmt
	table      string
	columns    map[string]string
	logger     *slog.Logger
}

func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	cfg.Logger = cfg.Logger.With("exporter", "PostgreSQL")
	db, err := sql.Open("postgres", cfg.PSQLInfo)
	if err != nil {
		return nil, err
	}
	if len(cfg.Columns) == 0 {
		cfg.Columns = sensor.DefaultColumnMap
	}
	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Using columns", slog.Any("columns", cfg.Columns))
	q, err := psql.BuildInsertQuery(cfg.Table, cfg.Columns)
	if err != nil {
		return nil, err
	}
	cfg.Logger.LogAttrs(ctx, slog.LevelDebug, "Preparing insert statement", slog.String("query", q))
	insertStmt, err := db.PrepareContext(ctx, q)
	if err != nil {
		return nil, err
	}
	return &postgresExporter{
		db:         db,
		insertStmt: insertStmt,
		table:      cfg.Table,
		columns:    cfg.Columns,
		logger:     cfg.Logger,
	}, nil
}

func (t *postgresExporter) Name() string {
	return "PostgreSQL"
}

func (t *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
	_, err := t.insertStmt.ExecContext(ctx, psql.BuildQueryArguments(t.columns, data)...)
	return err
}

func (t *postgresExporter) Close() error {
	return errors.Join(t.insertStmt.Close(), t.db.Close())
}
