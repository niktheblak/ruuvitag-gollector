//go:build postgres

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"

	_ "github.com/lib/pq"
)

const SchemaTmpl = `CREATE TABLE %s (
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
  measurement_number INTEGER
)`

type postgresExporter struct {
	db         *sql.DB
	insertStmt *sql.Stmt
}

func New(ctx context.Context, connStr, table string) (exporter.Exporter, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	insertStmt, err := db.PrepareContext(ctx, fmt.Sprintf(`
INSERT INTO %s (
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, table))
	if err != nil {
		return nil, err
	}
	return &postgresExporter{
		db:         db,
		insertStmt: insertStmt,
	}, nil
}

func (p *postgresExporter) Name() string {
	return "Postgres"
}

func (p *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
	_, err := p.insertStmt.ExecContext(ctx, data.Addr, data.Name, data.Timestamp, data.Temperature, data.Humidity, data.Pressure, data.AccelerationX, data.AccelerationY, data.AccelerationZ, data.MovementCounter, data.BatteryVoltage, data.MeasurementNumber)
	return err
}

func (p *postgresExporter) Close() error {
	p.insertStmt.Close()
	return p.db.Close()
}
