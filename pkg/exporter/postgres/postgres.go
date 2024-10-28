//go:build postgres

package postgres

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/niktheblak/ruuvitag-common/pkg/psql"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

type postgresExporter struct {
	name       string
	dbpool     *pgxpool.Pool
	connString string
	query      string
	columns    map[string]string
	logger     *slog.Logger
}

func New(ctx context.Context, name string, cfg Config) (exporter.Exporter, error) {
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
	dbpool, err := pgxpool.New(ctx, cfg.ConnString)
	if err != nil {
		return nil, err
	}
	q, err := psql.BuildInsertQuery(cfg.Table, cfg.Columns)
	if err != nil {
		return nil, err
	}
	cfg.Logger.LogAttrs(ctx, slog.LevelDebug, "Using insert query", slog.String("query", q))
	e := &postgresExporter{
		name:       name,
		dbpool:     dbpool,
		connString: cfg.ConnString,
		query:      q,
		columns:    cfg.Columns,
		logger:     cfg.Logger,
	}
	return e, nil
}

func (t *postgresExporter) Name() string {
	return t.name
}

func (t *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
	args := psql.BuildQueryArguments(t.columns, data)
	_, err := t.dbpool.Exec(ctx, t.query, args...)
	if err != nil {
		return err
	}
	return nil
}

func (t *postgresExporter) Close() error {
	t.dbpool.Close()
	return nil
}
