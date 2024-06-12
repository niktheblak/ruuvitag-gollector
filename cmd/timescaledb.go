//go:build timescaledb

package cmd

import (
	"context"
	"log/slog"

	"github.com/spf13/viper"

	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter/timescaledb"
	"github.com/niktheblak/ruuvitag-gollector/pkg/psql"
)

func init() {
	rootCmd.PersistentFlags().Bool("timescaledb.enabled", false, "enable TimescaleDB exporter")
	AddPsqlFlags(rootCmd.PersistentFlags(), "timescaledb")
}

func addTimescaleDBExporter(exporters *[]exporter.Exporter) error {
	ctx := context.Background()
	psqlInfo, err := CreatePsqlInfoString("timescaledb")
	if err != nil {
		return err
	}
	table := viper.GetString("timescaledb.table")
	logger.LogAttrs(ctx, slog.LevelInfo, "Connecting to TimescaleDB", slog.String("conn_str", psql.RemovePassword(psqlInfo)), slog.String("table", table))
	exp, err := timescaledb.New(ctx, psqlInfo, table, logger)
	if err != nil {
		return err
	}
	*exporters = append(*exporters, exp)
	return nil
}
