//go:build influxdb

package influxdb

import (
	"context"
	"crypto/tls"
	"fmt"
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

type pointWriter interface {
	WritePoint(ctx context.Context, point *write.Point) error
}

type influxdbExporter struct {
	pointWriter
	io.Closer
	client      influxdb2.Client
	measurement string
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

func New(cfg Config, logger *slog.Logger) (exporter.Exporter, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("InfluxDB host must be specified")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket must be specified")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("token must be specified")
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
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
				logger:      logger,
			},
			writeAPI: client.WriteAPI(cfg.Org, bucket),
		}
		exp.influxdbExporter.pointWriter = exp
		exp.Closer = exp
		exp.writeAPI.SetWriteFailedCallback(func(batch string, error http2.Error, retryAttempts uint) bool {
			logger.Error("Failed to write batch to InfluxDB", slog.String("batch", batch), slog.String("error", error.Error()), slog.Int("retry attempts", int(retryAttempts)))
			return retryAttempts <= 3
		})
		return exp, nil
	} else {
		exp := &blockingInfluxdbExporter{
			influxdbExporter: influxdbExporter{
				client:      client,
				measurement: cfg.Measurement,
				logger:      logger,
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
	point := influxdb2.NewPoint(e.measurement, map[string]string{
		"mac":  strings.ToUpper(data.Addr),
		"name": data.Name,
	}, map[string]interface{}{
		"temperature":        data.Temperature,
		"humidity":           data.Humidity,
		"dew_point":          data.DewPoint,
		"pressure":           data.Pressure,
		"battery_voltage":    data.BatteryVoltage,
		"tx_power":           data.TxPower,
		"acceleration_x":     data.AccelerationX,
		"acceleration_y":     data.AccelerationY,
		"acceleration_z":     data.AccelerationZ,
		"movement_counter":   data.MovementCounter,
		"measurement_number": data.MeasurementNumber,
	}, data.Timestamp)
	return e.WritePoint(ctx, point)
}

func (e *influxdbExporter) Close() error {
	return e.Close()
}
