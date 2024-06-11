//go:build timescaledb

package timescaledb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

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

	createExtensionStatement = `CREATE EXTENSION IF NOT EXISTS timescaledb`

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
	db    *sql.DB
	table string
}

func New(psqlInfo, table string) (exporter.Exporter, error) {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	return &timescaleDBExporter{
		db:    db,
		table: table,
	}, nil
}

func (t *timescaleDBExporter) Name() string {
	return "TimescaleDB"
}

func (t *timescaleDBExporter) Export(ctx context.Context, data sensor.Data) error {
	q := fmt.Sprintf(insertTemplate, t.table)
	_, err := t.db.ExecContext(
		ctx,
		q,
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
	_, err := t.db.ExecContext(ctx, createExtensionStatement, t.table)
	if err != nil {
		return err
	}
	_, err = t.db.ExecContext(ctx, fmt.Sprintf(createSchemaTemplate, t.table))
	if err != nil {
		return err
	}
	_, err = t.db.ExecContext(ctx, fmt.Sprintf(createHyperTableTemplate, t.table))
	if err != nil {
		return err
	}
	return nil
}

func (t *timescaleDBExporter) Close() error {
	return t.db.Close()
}
