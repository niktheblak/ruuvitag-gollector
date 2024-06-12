//go:build timescaledb

package timescaledb

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
        time,
        mac,
        name,
        temperature,
        humidity,
        pressure,
        acceleration_x,
        acceleration_y,
        acceleration_z,
        movement_counter,
        battery,
        measurement_number,
        dew_point,
        battery_voltage,
        tx_power
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	createExtensionStmt = `CREATE EXTENSION IF NOT EXISTS timescaledb`

	createSchemaTemplate = `CREATE TABLE %s (
		time TIMESTAMPTZ NOT NULL,
		mac MACADDR NOT NULL,
		name TEXT,
		temperature DOUBLE PRECISION,
		humidity DOUBLE PRECISION,
		pressure DOUBLE PRECISION,
		acceleration_x INTEGER,
		acceleration_y INTEGER,
		acceleration_z INTEGER,
		movement_counter INTEGER,
		battery DOUBLE PRECISION,
		measurement_number INTEGER,
		dew_point DOUBLE PRECISION,
		battery_voltage DOUBLE PRECISION,
		tx_power INTEGER)`

	createHyperTableTemplate = `SELECT create_hypertable('%s', by_range('time'))`
)

type timescaleDBExporter struct {
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
	return &timescaleDBExporter{
		db:         db,
		insertStmt: insertStmt,
		table:      table,
		logger:     logger,
	}, nil
}

func (t *timescaleDBExporter) Name() string {
	return "TimescaleDB"
}

func (t *timescaleDBExporter) Export(ctx context.Context, data sensor.Data) error {
	_, err := t.insertStmt.ExecContext(
		ctx,
		data.Timestamp,
		data.Addr,
		data.Name,
		data.Temperature,
		data.Humidity,
		data.Pressure,
		data.AccelerationX,
		data.AccelerationY,
		data.AccelerationZ,
		data.MovementCounter,
		data.BatteryVoltage,
		data.MeasurementNumber,
		data.DewPoint,
		data.BatteryVoltage,
		data.TxPower,
	)
	return err
}

func (t *timescaleDBExporter) InitSchema(ctx context.Context) error {
	t.logger.LogAttrs(ctx, slog.LevelDebug, "Creating extension", slog.String("query", psql.TrimQuery(createExtensionStmt)), slog.String("table", t.table))
	_, err := t.db.ExecContext(ctx, createExtensionStmt, t.table)
	if err != nil {
		return err
	}
	q := fmt.Sprintf(createSchemaTemplate, t.table)
	t.logger.LogAttrs(ctx, slog.LevelDebug, "Creating schema", slog.String("query", psql.TrimQuery(q)))
	_, err = t.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	q = fmt.Sprintf(createHyperTableTemplate, t.table)
	t.logger.LogAttrs(ctx, slog.LevelDebug, "Creating hypertable", slog.String("query", psql.TrimQuery(q)))
	_, err = t.db.ExecContext(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (t *timescaleDBExporter) Close() error {
	return errors.Join(t.insertStmt.Close(), t.db.Close())
}
