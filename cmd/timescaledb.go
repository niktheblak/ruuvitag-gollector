//go:build timescaledb

package cmd

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/timescaledb"
)

func init() {
	rootCmd.PersistentFlags().Bool("timescaledb.enabled", false, "Store measurements to TimescaleDB")
	rootCmd.PersistentFlags().String("timescaledb.host", "", "TimescaleDB host")
	rootCmd.PersistentFlags().Int("timescaledb.port", 0, "TimescaleDB port")
	rootCmd.PersistentFlags().String("timescaledb.username", "", "TimescaleDB username")
	rootCmd.PersistentFlags().String("timescaledb.password", "", "TimescaleDB username")
	rootCmd.PersistentFlags().String("timescaledb.database", "", "TimescaleDB database")
	rootCmd.PersistentFlags().String("timescaledb.table", "", "TimescaleDB table")

	viper.SetDefault("timescaledb.port", "5432")
}

func addTimescaleDBExporter(exporters *[]exporter.Exporter) error {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		viper.GetString("timescaledb.host"),
		viper.GetInt("timescaledb.port"),
		viper.GetString("timescaledb.username"),
		viper.GetString("timescaledb.password"),
		viper.GetString("timescaledb.database"),
	)
	exp, err := timescaledb.New(psqlInfo, viper.GetString("timescaledb.table"))
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
