//go:build postgres

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"

	_ "github.com/lib/pq"
)

const (
	insertTemplate = `INSERT INTO %s (
		mac,
		name,
		ts,
		temperature,
		humidity,
		pressure,
		acceleration_x,
		acceleration_y,
		acceleration_z,
		movement_counter,
		battery,
		measurement_number
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	createSchemaTemplate = `CREATE TABLE %s (
		id BIGSERIAL PRIMARY KEY,
		mac MACADDR NOT NULL,
		name TEXT,
		ts TIMESTAMP NOT NULL,
		temperature REAL,
		humidity REAL,
		pressure REAL,
		acceleration_x INTEGER,
		acceleration_y INTEGER,
		acceleration_z INTEGER,
		movement_counter INTEGER,
		battery REAL,
		measurement_number INTEGER)`
)

type postgresExporter struct {
	db         *sql.DB
	insertStmt *sql.Stmt
	table      string
	logger     *slog.Logger
}

func New(ctx context.Context, psqlInfo, table string, logger *slog.Logger) (exporter.Exporter, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf(insertTemplate, table)
	logger.LogAttrs(ctx, slog.LevelDebug, "Preparing insert statement", slog.String("query", psql.TrimQuery(q)))
	insertStmt, err := db.PrepareContext(ctx, q)
	if err != nil {
		return nil, err
	}
	return &postgresExporter{
		db:         db,
		insertStmt: insertStmt,
		table:      table,
		logger:     logger,
	}, nil
}

func (p *postgresExporter) Name() string {
	return "Postgres"
}

func (p *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
	_, err := p.insertStmt.ExecContext(
		ctx,
		data.Addr,
		data.Name,
		data.Timestamp,
		data.Temperature,
		data.Humidity,
		data.Pressure,
		data.AccelerationX,
		data.AccelerationY,
		data.AccelerationZ,
		data.MovementCounter,
		data.BatteryVoltage,
		data.MeasurementNumber,
	)
	return err
}

func (p *postgresExporter) InitSchema(ctx context.Context) error {
	q := fmt.Sprintf(createSchemaTemplate, p.table)
	p.logger.LogAttrs(ctx, slog.LevelDebug, "Creating schema", slog.String("query", psql.TrimQuery(q)))
	_, err := p.db.ExecContext(ctx, q)
	return err
}

func (p *postgresExporter) Close() error {
	return errors.Join(p.insertStmt.Close(), p.db.Close())
}
