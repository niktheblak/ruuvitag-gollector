//go:build influxdb

package cmd

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
)

func init() {
	rootCmd.PersistentFlags().Bool("influxdb.enabled", false, "Store measurements to InfluxDB")
	rootCmd.PersistentFlags().String("influxdb.addr", "http://localhost:8086", "InfluxDB address with protocol, host and port")
	rootCmd.PersistentFlags().String("influxdb.org", "", "InfluxDB organization")
	rootCmd.PersistentFlags().String("influxdb.bucket", "", "InfluxDB bucket")
	rootCmd.PersistentFlags().String("influxdb.database", "", "InfluxDB database (1.x)")
	rootCmd.PersistentFlags().String("influxdb.measurement", "", "InfluxDB measurement name")
	rootCmd.PersistentFlags().String("influxdb.token", "", "InfluxDB token")
	rootCmd.PersistentFlags().String("influxdb.username", "", "InfluxDB username (1.x)")
	rootCmd.PersistentFlags().String("influxdb.password", "", "InfluxDB password (1.x)")
}

func addInfluxDBExporter(exporters *[]exporter.Exporter) error {
	addr := viper.GetString("influxdb.addr")
	if addr == "" {
		return fmt.Errorf("InfluxDB address must be specified")
	}
	cfg := influxdb.Config{
		Addr:        addr,
		Org:         viper.GetString("influxdb.org"),
		Bucket:      viper.GetString("influxdb.bucket"),
		Database:    viper.GetString("influxdb.database"),
		Measurement: viper.GetString("influxdb.measurement"),
		Token:       viper.GetString("influxdb.token"),
		Username:    viper.GetString("influxdb.username"),
		Password:    viper.GetString("influxdb.password"),
	}
	logger.Info("Connecting to InfluxDB", zap.String("addr", cfg.Addr), zap.String("org", cfg.Org), zap.String("bucket", cfg.Bucket), zap.String("database", cfg.Database), zap.String("measurement", cfg.Measurement))
	influx := influxdb.New(cfg)
	*exporters = append(*exporters, influx)
	return nil
}
