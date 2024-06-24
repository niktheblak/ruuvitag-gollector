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
        %s,
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
)

type postgresExporter struct {
	db         *sql.DB
	insertStmt *sql.Stmt
	table      string
	logger     *slog.Logger
}

func New(ctx context.Context, psqlInfo, table, timeColumn string, logger *slog.Logger) (exporter.Exporter, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf(insertTemplate, table, timeColumn)
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

func (t *postgresExporter) Name() string {
	return "PostgreSQL"
}

func (t *postgresExporter) Export(ctx context.Context, data sensor.Data) error {
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

func (t *postgresExporter) Close() error {
	return errors.Join(t.insertStmt.Close(), t.db.Close())
}
