//go:build postgres

package postgres

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"
)

type postgresExporter struct {
	conn    *pgx.Conn
	query   string
	columns map[string]string
	logger  *slog.Logger
}

func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	cfg.Logger = cfg.Logger.With("exporter", "PostgreSQL")
	conn, err := pgx.Connect(ctx, cfg.ConnString)
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
	return &postgresExporter{
		conn:    conn,
		query:   q,
		columns: cfg.Columns,
		logger:  cfg.Logger,
	}, nil
}

func (t *postgresExporter) Name() string {
	return "PostgreSQL"
}

func (t *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
	_, err := t.conn.Exec(ctx, t.query, psql.BuildQueryArguments(t.columns, data)...)
	return err
}

func (t *postgresExporter) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.conn.Close(ctx)
}
