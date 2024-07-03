//go:build influxdb

package influxdb

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	http2 "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

var defaultColumns = map[string]string{
	"temperature":        "temperature",
	"humidity":           "humidity",
	"dew_point":          "dew_point",
	"pressure":           "pressure",
	"battery_voltage":    "battery_voltage",
	"tx_power":           "tx_power",
	"acceleration_x":     "acceleration_x",
	"acceleration_y":     "acceleration_y",
	"acceleration_z":     "acceleration_z",
	"movement_counter":   "movement_counter",
	"measurement_number": "measurement_number",
}

type pointWriter interface {
	WritePoint(ctx context.Context, point *write.Point) error
}

type influxdbExporter struct {
	pointWriter
	io.Closer
	client      influxdb2.Client
	measurement string
	columns     map[string]string
	logger      *slog.Logger
}

type asyncInfluxdbExporter struct {
	influxdbExporter
	writeAPI api.WriteAPI
}

func (e *asyncInfluxdbExporter) WritePoint(ctx context.Context, point *write.Point) error {
	e.writeAPI.WritePoint(point)
	return nil
}

func (e *asyncInfluxdbExporter) Close() error {
	e.writeAPI.Flush()
	e.client.Close()
	return nil
}

type blockingInfluxdbExporter struct {
	influxdbExporter
	writeAPIBlocking api.WriteAPIBlocking
}

func (e *blockingInfluxdbExporter) WritePoint(ctx context.Context, point *write.Point) error {
	return e.writeAPIBlocking.WritePoint(ctx, point)
}

func (e *blockingInfluxdbExporter) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err := e.writeAPIBlocking.Flush(ctx)
	cancel()
	e.client.Close()
	return err
}

func New(cfg Config) (exporter.Exporter, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	if len(cfg.Columns) == 0 {
		cfg.Columns = defaultColumns
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	cfg.Logger = cfg.Logger.With("exporter", "InfluxDB")
	opts := influxdb2.DefaultOptions()
	opts.SetUseGZip(true)
	opts.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	if cfg.BatchSize > 0 {
		opts.SetBatchSize(uint(cfg.BatchSize))
	}
	if cfg.FlushInterval > 0 {
		opts.SetFlushInterval(uint(cfg.FlushInterval / time.Millisecond))
	}
	client := influxdb2.NewClientWithOptions(cfg.Addr, cfg.Token, opts)
	bucket := cfg.Bucket
	if cfg.Async {
		exp := &asyncInfluxdbExporter{
			influxdbExporter: influxdbExporter{
				client:      client,
				measurement: cfg.Measurement,
				columns:     cfg.Columns,
				logger:      cfg.Logger,
			},
			writeAPI: client.WriteAPI(cfg.Org, bucket),
		}
		exp.influxdbExporter.pointWriter = exp
		exp.Closer = exp
		exp.writeAPI.SetWriteFailedCallback(func(batch string, error http2.Error, retryAttempts uint) bool {
			exp.logger.LogAttrs(nil, slog.LevelError, "Failed to write batch to InfluxDB", slog.String("batch", batch), slog.String("error", error.Error()), slog.Int("retry attempts", int(retryAttempts)))
			return retryAttempts <= 3
		})
		return exp, nil
	} else {
		exp := &blockingInfluxdbExporter{
			influxdbExporter: influxdbExporter{
				client:      client,
				measurement: cfg.Measurement,
				columns:     cfg.Columns,
				logger:      cfg.Logger,
			},
			writeAPIBlocking: client.WriteAPIBlocking(cfg.Org, bucket),
		}
		exp.influxdbExporter.pointWriter = exp
		exp.Closer = exp
		return exp, nil
	}
}

func (e *influxdbExporter) Name() string {
	return "InfluxDB"
}

func (e *influxdbExporter) Export(ctx context.Context, data sensor.Data) error {
	fields := make(map[string]interface{})
	for dc := range defaultColumns {
		cn, ok := e.columns[dc]
		if !ok {
			continue
		}
		switch dc {
		case "temperature":
			fields[cn] = data.Temperature
		case "humidity":
			fields[cn] = data.Humidity
		case "dew_point":
			fields[cn] = data.DewPoint
		case "pressure":
			fields[cn] = data.Pressure
		case "battery_voltage":
			fields[cn] = data.BatteryVoltage
		case "tx_power":
			fields[cn] = data.TxPower
		case "acceleration_x":
			fields[cn] = data.AccelerationX
		case "acceleration_y":
			fields[cn] = data.AccelerationY
		case "acceleration_z":
			fields[cn] = data.AccelerationZ
		case "movement_counter":
			fields[cn] = data.MovementCounter
		case "measurement_number":
			fields[cn] = data.MeasurementNumber
		}
	}
	point := influxdb2.NewPoint(e.measurement, map[string]string{
		"mac":  strings.ToUpper(data.Addr),
		"name": data.Name,
	}, fields, data.Timestamp)
	return e.WritePoint(ctx, point)
}

func (e *influxdbExporter) Close() error {
	return e.Close()
}
