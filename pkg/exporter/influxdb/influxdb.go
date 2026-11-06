//go:build influxdb

package influxdb

import (
	"context"
	"log/slog"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	influxhttp "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"

	"github.com/niktheblak/ruuvitag-common/pkg/columnmap"
	"github.com/niktheblak/ruuvitag-common/pkg/sensor"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
)

const closeTimeout = 5 * time.Second

// pointWriter abstracts over the blocking and asynchronous InfluxDB write APIs.
type pointWriter interface {
	WritePoint(ctx context.Context, point *write.Point) error
	Flush(ctx context.Context) error
}

type asyncWriter struct {
	writeAPI api.WriteAPI
}

func (w *asyncWriter) WritePoint(_ context.Context, point *write.Point) error {
	w.writeAPI.WritePoint(point)
	return nil
}

func (w *asyncWriter) Flush(_ context.Context) error {
	w.writeAPI.Flush()
	return nil
}

type blockingWriter struct {
	writeAPI api.WriteAPIBlocking
}

func (w *blockingWriter) WritePoint(ctx context.Context, point *write.Point) error {
	return w.writeAPI.WritePoint(ctx, point)
}

func (w *blockingWriter) Flush(ctx context.Context) error {
	return w.writeAPI.Flush(ctx)
}

type influxdbExporter struct {
	name        string
	client      influxdb2.Client
	writer      pointWriter
	measurement string
	columns     map[string]string
	logger      *slog.Logger
}

func New(ctx context.Context, name string, cfg Config) (exporter.Exporter, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	if cfg.Measurement == "" {
		cfg.Measurement = DefaultMeasurement
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.DiscardHandler)
	}
	cfg.Logger = cfg.Logger.With("exporter", "InfluxDB")
	opts := influxdb2.DefaultOptions()
	opts.SetUseGZip(true)
	if cfg.BatchSize > 0 {
		opts.SetBatchSize(uint(cfg.BatchSize))
	}
	if cfg.FlushInterval > 0 {
		opts.SetFlushInterval(uint(cfg.FlushInterval / time.Millisecond))
	}
	cfg.Logger.LogAttrs(ctx, slog.LevelDebug, "Creating InfluxDB client",
		slog.String("addr", cfg.Addr),
		slog.String("org", cfg.Org),
		slog.String("bucket", cfg.Bucket),
		slog.String("measurement", cfg.Measurement),
		slog.Bool("async", cfg.Async),
	)
	client := influxdb2.NewClientWithOptions(cfg.Addr, cfg.Token, opts)
	e := &influxdbExporter{
		name:        name,
		client:      client,
		measurement: cfg.Measurement,
		columns:     cfg.Columns,
		logger:      cfg.Logger,
	}
	if cfg.Async {
		writeAPI := client.WriteAPI(cfg.Org, cfg.Bucket)
		writeAPI.SetWriteFailedCallback(func(batch string, err influxhttp.Error, retryAttempts uint) bool {
			e.logger.LogAttrs(context.Background(), slog.LevelError, "Failed to write batch to InfluxDB",
				slog.String("batch", batch),
				slog.String("error", err.Error()),
				slog.Int("retry_attempts", int(retryAttempts)),
			)
			return retryAttempts <= 3
		})
		e.writer = &asyncWriter{writeAPI: writeAPI}
	} else {
		e.writer = &blockingWriter{writeAPI: client.WriteAPIBlocking(cfg.Org, cfg.Bucket)}
	}
	return e, nil
}

func (e *influxdbExporter) Name() string {
	return e.name
}

func (e *influxdbExporter) Export(ctx context.Context, data sensor.Data) error {
	fields := columnmap.Transform(e.columns, data)
	tags := make(map[string]string)
	// time is the point timestamp and mac/name are tags, so they are removed from fields
	delete(fields, e.columns["time"])
	if mac, ok := e.columns["mac"]; ok {
		delete(fields, mac)
		tags[mac] = strings.ToUpper(data.Addr)
	}
	if name, ok := e.columns["name"]; ok {
		delete(fields, name)
		tags[name] = data.Name
	}
	point := influxdb2.NewPoint(e.measurement, tags, fields, data.Timestamp)
	e.logger.LogAttrs(ctx, slog.LevelDebug, "Writing point", slog.String("measurement", e.measurement), slog.Any("tags", tags), slog.Any("fields", fields))
	return e.writer.WritePoint(ctx, point)
}

func (e *influxdbExporter) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), closeTimeout)
	defer cancel()
	err := e.writer.Flush(ctx)
	e.client.Close()
	return err
}
