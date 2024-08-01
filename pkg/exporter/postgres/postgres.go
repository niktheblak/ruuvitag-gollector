//go:build postgres

package postgres

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"
)

type postgresExporter struct {
	conn       *pgx.Conn
	connString string
	query      string
	columns    map[string]string
	logger     *slog.Logger
}

func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	cfg.Logger = cfg.Logger.With("exporter", "PostgreSQL")
	if len(cfg.Columns) == 0 {
		cfg.Columns = sensor.DefaultColumnMap
	}
	cfg.Logger.LogAttrs(ctx, slog.LevelInfo, "Using columns", slog.Any("columns", cfg.Columns))
	q, err := psql.BuildInsertQuery(cfg.Table, cfg.Columns)
	if err != nil {
		return nil, err
	}
	cfg.Logger.LogAttrs(ctx, slog.LevelDebug, "Using insert query", slog.String("query", q))
	e := &postgresExporter{
		connString: cfg.ConnString,
		query:      q,
		columns:    cfg.Columns,
		logger:     cfg.Logger,
	}
	return e, e.reconnect(ctx)
}

func (t *postgresExporter) Name() string {
	return "PostgreSQL"
}

func (t *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
	if t.conn == nil || t.conn.IsClosed() {
		if err := t.reconnect(ctx); err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
	}
	_, err := t.conn.Exec(ctx, t.query, psql.BuildQueryArguments(t.columns, data)...)
	return err
}

func (t *postgresExporter) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.conn.Close(ctx)
}

func (t *postgresExporter) reconnect(ctx context.Context) error {
	if t.conn != nil {
		if err := t.conn.Close(ctx); err != nil {
			t.logger.LogAttrs(ctx, slog.LevelWarn, "Error while closing connection", slog.String("error", err.Error()))
		}
	}
	t.conn = nil
	conn, err := pgx.Connect(ctx, t.connString)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}
