// +build influxdb

package cmd

import (
	"fmt"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/influxdb"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.PersistentFlags().Bool("influxdb.enabled", false, "Store measurements to InfluxDB")
	rootCmd.PersistentFlags().String("influxdb.addr", "http://localhost:8086", "InfluxDB address with protocol, host and port")
	rootCmd.PersistentFlags().String("influxdb.database", "", "InfluxDB database to use")
	rootCmd.PersistentFlags().String("influxdb.measurement", "", "InfluxDB measurement name")
	rootCmd.PersistentFlags().String("influxdb.username", "", "InfluxDB username")
	rootCmd.PersistentFlags().String("influxdb.password", "", "InfluxDB password")
}

func addInfluxDBExporter(exporters *[]exporter.Exporter) error {
	addr := viper.GetString("influxdb.addr")
	if addr == "" {
		return fmt.Errorf("InfluxDB address must be specified")
	}
	influx, err := influxdb.New(influxdb.Config{
		Addr:        addr,
		Database:    viper.GetString("influxdb.database"),
		Measurement: viper.GetString("influxdb.measurement"),
		Username:    viper.GetString("influxdb.username"),
		Password:    viper.GetString("influxdb.password"),
	})
	if err != nil {
		return fmt.Errorf("failed to create InfluxDB exporter: %w", err)
	}
	*exporters = append(*exporters, influx)
	return nil
}
