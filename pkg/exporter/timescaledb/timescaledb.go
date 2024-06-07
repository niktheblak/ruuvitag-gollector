//go:build timescaledb

package timescaledb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

	_ "github.com/lib/pq"
)

const (
	CreateExtensionTmpl = `CREATE EXTENSION IF NOT EXISTS timescaledb`
	CreateSchemaTmpl    = `CREATE TABLE %s (
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
		measurement_number INTEGER)`
	CreateHyperTableTmpl = `SELECT create_hypertable('%s', by_range('time'))`
)

type timescaleDBExporter struct {
	db         *sql.DB
	insertStmt *sql.Stmt
}

func New(ctx context.Context, connStr, table string) (exporter.Exporter, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	insertStmt, err := db.PrepareContext(ctx, fmt.Sprintf(`INSERT INTO %s (mac,
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
        measurement_number)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, table))
	if err != nil {
		return nil, err
	}
	return &timescaleDBExporter{
		db:         db,
		insertStmt: insertStmt,
	}, nil
}

func (t *timescaleDBExporter) Name() string {
	return "TimescaleDB"
}

func (t *timescaleDBExporter) Export(ctx context.Context, data sensor.Data) error {
	_, err := t.insertStmt.ExecContext(ctx, data.Addr, data.Name, data.Timestamp, data.Temperature, data.Humidity, data.Pressure, data.AccelerationX, data.AccelerationY, data.AccelerationZ, data.MovementCounter, data.BatteryVoltage, data.MeasurementNumber)
	return err
}

func (t *timescaleDBExporter) Close() error {
	return errors.Join(t.insertStmt.Close(), t.db.Close())
}
