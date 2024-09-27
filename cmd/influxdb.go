//go:build influxdb

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cast"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
)

func createInfluxDBExporter(columns map[string]string, cfg map[string]any) (exporter.Exporter, error) {
	addr := cast.ToString(cfg["addr"])
	if addr == "" {
		return nil, fmt.Errorf("InfluxDB address must be specified")
	}
	influxCfg := influxdb.Config{
		Addr:          addr,
		Org:           cast.ToString(cfg["org"]),
		Bucket:        cast.ToString(cfg["bucket"]),
		Measurement:   cast.ToString(cfg["measurement"]),
		Token:         cast.ToString(cfg["token"]),
		Async:         cast.ToBool(cfg["async"]),
		BatchSize:     cast.ToInt(cfg["batch_size"]),
		FlushInterval: cast.ToDuration(cfg["flush_interval"]),
		Columns:       columns,
		Logger:        logger,
	}
	logger.LogAttrs(
		nil,
		slog.LevelInfo,
		"Connecting to InfluxDB",
		slog.Any("config", influxCfg),
		slog.Any("columns", columns),
	)
	return influxdb.New(influxCfg)
}
