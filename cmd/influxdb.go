//go:build influxdb

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
)

func init() {
	rootCmd.PersistentFlags().Bool("influxdb.enabled", false, "Store measurements to InfluxDB")
	rootCmd.PersistentFlags().String("influxdb.addr", "http://localhost:8086", "InfluxDB address with protocol, host and port")
	rootCmd.PersistentFlags().String("influxdb.org", "", "InfluxDB organization")
	rootCmd.PersistentFlags().String("influxdb.bucket", "", "InfluxDB bucket")
	rootCmd.PersistentFlags().String("influxdb.measurement", "", "InfluxDB measurement name")
	rootCmd.PersistentFlags().String("influxdb.token", "", "InfluxDB token")
	rootCmd.PersistentFlags().Bool("influxdb.async", false, "Write measurements asynchronously")
	rootCmd.PersistentFlags().Int("influxdb.batch_size", 0, "InfluxDB client batch size")
	rootCmd.PersistentFlags().Duration("influxdb.flush_interval", 0, "InfluxDB client flush interval")
}

func addInfluxDBExporter(exporters *[]exporter.Exporter, columns map[string]string) error {
	addr := viper.GetString("influxdb.addr")
	if addr == "" {
		return fmt.Errorf("InfluxDB address must be specified")
	}
	cfg := influxdb.Config{
		Addr:          addr,
		Org:           viper.GetString("influxdb.org"),
		Bucket:        viper.GetString("influxdb.bucket"),
		Measurement:   viper.GetString("influxdb.measurement"),
		Token:         viper.GetString("influxdb.token"),
		Async:         viper.GetBool("influxdb.async"),
		BatchSize:     viper.GetInt("influxdb.batch_size"),
		FlushInterval: viper.GetDuration("influxdb.flush_interval"),
		Columns:       columns,
		Logger:        logger,
	}
	logger.LogAttrs(nil, slog.LevelInfo, "Connecting to InfluxDB", slog.String("addr", cfg.Addr), slog.String("org", cfg.Org), slog.String("bucket", cfg.Bucket), slog.String("measurement", cfg.Measurement), slog.Bool("async", cfg.Async), slog.Int("batch_size", cfg.BatchSize), slog.Duration("flush_interval", cfg.FlushInterval), slog.Any("columns", columns))
	influx, err := influxdb.New(cfg)
	if err != nil {
		return err
	}
	*exporters = append(*exporters, influx)
	return nil
}
