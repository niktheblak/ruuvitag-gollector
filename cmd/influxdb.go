//go:build influxdb

package cmd

import (
	"context"
	"log/slog"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
)

func createInfluxDBExporter(name string, columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), exporterInitTimeout)
	defer cancel()
	flushInterval, err := configDuration(cfg["flush_interval"])
	if err != nil {
		return nil, err
	}
	influxCfg := influxdb.Config{
		Addr:          cast.ToString(cfg["addr"]),
		Org:           cast.ToString(cfg["org"]),
		Bucket:        cast.ToString(cfg["bucket"]),
		Measurement:   cast.ToString(cfg["measurement"]),
		Token:         cast.ToString(cfg["token"]),
		Async:         cast.ToBool(cfg["async"]),
		BatchSize:     cast.ToInt(cfg["batch_size"]),
		FlushInterval: flushInterval,
		Columns:       columns,
		Logger:        logger,
	}
	logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to InfluxDB",
		slog.String("addr", influxCfg.Addr),
		slog.String("org", influxCfg.Org),
		slog.String("bucket", influxCfg.Bucket),
		slog.String("measurement", influxCfg.Measurement),
		slog.Bool("async", influxCfg.Async),
		slog.Int("batch_size", influxCfg.BatchSize),
		slog.Duration("flush_interval", influxCfg.FlushInterval),
	)
	return influxdb.New(ctx, name, influxCfg)
}
