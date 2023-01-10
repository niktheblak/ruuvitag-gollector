//go:build influxdb

package cmd

import (
	"fmt"

	"github.com/spf13/viper"

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
	influx := influxdb.New(influxdb.Config{
		Addr:        addr,
		Org:         viper.GetString("influxdb.org"),
		Bucket:      viper.GetString("influxdb.bucket"),
		Database:    viper.GetString("influxdb.database"),
		Measurement: viper.GetString("influxdb.measurement"),
		Token:       viper.GetString("influxdb.token"),
		Username:    viper.GetString("influxdb.username"),
		Password:    viper.GetString("influxdb.password"),
	})
	*exporters = append(*exporters, influx)
	return nil
}
