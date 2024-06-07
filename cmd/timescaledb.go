//go:build postgres

package cmd

import (
	"context"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/timescaledb"
)

func init() {
	rootCmd.PersistentFlags().Bool("timescaledb.enabled", false, "Store measurements to TimescaleDB")
	rootCmd.PersistentFlags().String("timescaledb.conn", "", "TimescaleDB connection string")
	rootCmd.PersistentFlags().String("timescaledb.table", "", "TimescaleDB table")
}

func addTimescaleDBExporter(exporters *[]exporter.Exporter) error {
	ctx := context.Background()
	connStr := viper.GetString("timescaledb.conn")
	table := viper.GetString("timescaledb.table")
	exp, err := timescaledb.New(ctx, connStr, table)
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
