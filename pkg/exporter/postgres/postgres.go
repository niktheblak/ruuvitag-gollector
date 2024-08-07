//go:build postgres

package postgres

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/niktheblak/ruuvitag-common/pkg/psql"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type postgresExporter struct {
	conn       *pgx.Conn
	connString string
	query      string
	columns    map[string]string
	logger     *slog.Logger
}

func New(ctx context.Context, cfg Config) (exporter.Exporter, error) {
	if cfg.ConnString == "" {
		return nil, fmt.Errorf("no connection string provided")
	}
	if len(cfg.Columns) == 0 {
		return nil, fmt.Errorf("columns must be non-empty")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	cfg.Logger = cfg.Logger.With("exporter", "PostgreSQL")
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
	// previous reconnect failed; attempt again
	if t.conn == nil {
		if err := t.reconnect(ctx); err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
	}
	args := psql.BuildQueryArguments(t.columns, data)
	_, err := t.conn.Exec(ctx, t.query, args...)
	if err != nil {
		if err.Error() == "conn closed" {
			// reconnect and retry
			if reconnectErr := t.reconnect(ctx); reconnectErr != nil {
				return fmt.Errorf("failed to reconnect: %w", reconnectErr)
			}
			_, err = t.conn.Exec(ctx, t.query, args...)
			return err
		} else {
			return err
		}
	}
	return nil
}

func (t *postgresExporter) Close() error {
	if t.conn == nil || t.conn.IsClosed() {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := t.conn.Close(ctx)
	t.conn = nil
	return err
}

func (t *postgresExporter) reconnect(ctx context.Context) error {
	if t.conn != nil {
		if err := t.conn.Close(ctx); err != nil {
			t.logger.LogAttrs(ctx, slog.LevelWarn, "Error while closing connection", slog.String("error", err.Error()))
		}
		t.conn = nil
	}
	conn, err := pgx.Connect(ctx, t.connString)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}
