//go:build postgres

package postgres

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/niktheblak/pgx-reconnect"
	"github.com/niktheblak/ruuvitag-common/pkg/psql"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type postgresExporter struct {
	conn       *pgxreconnect.ReConn
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
	conn, err := pgxreconnect.Connect(ctx, cfg.ConnString, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}
	q, err := psql.BuildInsertQuery(cfg.Table, cfg.Columns)
	if err != nil {
		return nil, err
	}
	cfg.Logger.LogAttrs(ctx, slog.LevelDebug, "Using insert query", slog.String("query", q))
	e := &postgresExporter{
		conn:       conn,
		connString: cfg.ConnString,
		query:      q,
		columns:    cfg.Columns,
		logger:     cfg.Logger,
	}
	return e, nil
}

func (t *postgresExporter) Name() string {
	return "PostgreSQL"
}

func (t *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
	args := psql.BuildQueryArguments(t.columns, data)
	_, err := t.conn.Exec(ctx, t.query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (t *postgresExporter) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := t.conn.Close(ctx)
	return err
}
