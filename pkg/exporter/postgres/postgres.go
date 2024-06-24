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

type Config struct {
	PSQLInfo string
	Table    string
	Columns  map[string]string
	Logger   *slog.Logger
}

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
	db, err := sql.Open("postgres", cfg.PSQLInfo)
	if err != nil {
		return nil, err
	}
	if len(cfg.Columns) == 0 {
		cfg.Columns = psql.DefaultColumns
	}
	q, err := psql.RenderInsertQuery(cfg.Table, cfg.Columns)
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
	_, err := t.insertStmt.ExecContext(ctx, psql.BuildQuery(t.columns, data)...)
	return err
}

func (t *postgresExporter) Close() error {
	return errors.Join(t.insertStmt.Close(), t.db.Close())
}
